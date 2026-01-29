package internal

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fixkme/gokit/framework/app"
	"github.com/fixkme/gokit/framework/core"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/wsg"
	"github.com/fixkme/othello/server/common"
	"github.com/fixkme/othello/server/common/env"
	"github.com/fixkme/othello/server/common/values"
	"github.com/fixkme/othello/server/pb/game"
	"github.com/fixkme/othello/server/pb/hall"
	"github.com/panjf2000/gnet/v2"
)

type GateServer struct {
	wsServer   *wsg.Server
	routerPool *_LoadBalanceImp
	name       string
	retired    bool // server是否Shutdown
}

func NewGateModule() app.Module {
	m := &GateServer{
		name: "gate",
	}
	return m
}

func (s *GateServer) OnInit() error {
	gnetOpt := gnet.Options{
		NumEventLoop: 4,
		LockOSThread: true,
		ReuseAddr:    true,
	}
	listenAddr := env.GetEnvStr(env.APP_GateListenAddr)
	opt := &wsg.ServerOptions{
		Options:          gnetOpt,
		Addr:             fmt.Sprintf("tcp4://%s", listenAddr), // "tcp4://:7070",
		HandshakeTimeout: 5000,
		OnHandshake:      func(conn *wsg.Conn, r *http.Request) error { return s.routerPool.OnHandshake(conn, r) },
		OnClientClose:    s.OnClientClose,
		OnServerShutdown: func(_ gnet.Engine) {
			s.routerPool.Stop()
			mlog.Infof("gate server shutdowned")
		},
	}
	s.wsServer = wsg.NewServer(opt)
	s.routerPool = newRouterPool(4, 1024)
	mlog.Infof("gate server listenAddr: (%s)", listenAddr)
	return nil
}

func (s *GateServer) Run() {
	s.routerPool.Start()
	if err := s.wsServer.Run(); err != nil {
		mlog.Errorf("GateServer Run error: %v", err)
		panic(err)
	}
	mlog.Infof("gate server exit run")
}

func (s *GateServer) Destroy() {
	mlog.Infof("gate server stop")
	s.retired = true
	if err := s.wsServer.Stop(context.Background()); err != nil {
		mlog.Errorf("wsServer Stop error:%v", err)
	}
	mlog.Infof("gate server stoped")
}

func (s *GateServer) Name() string {
	return s.name
}

func (s *GateServer) OnClientClose(conn *wsg.Conn, err error) {
	if s.retired {
		// 主动关闭ws server
		return
	}
	cli, ok := conn.GetSession().(*WsClient)
	if !ok {
		mlog.Infof("OnClientClose when handshake %v", conn.RemoteAddr().String())
		return
	}
	pid := cli.PlayerId
	mlog.Infof("player ws closed, acc:%s, pid:%d, addr:%s, reason:%v", cli.Account, pid, conn.RemoteAddr().String(), err)
	if pid > 0 {
		ClientMgr.RemoveClient(pid)
		meta := common.WarpMeta(cli.PlayerId, GateNodeId)
		// 通知hall玩家下线
		callErr := core.Rpc.AsyncCallWithoutResp(values.Service_Hall, &hall.CPlayerOffline{PlayerId: pid})
		if callErr != nil {
			mlog.Errorf("player %d call hall.PlayerOffline failed, %v", pid, callErr)
		}
		// 通知game玩家下线
		gameServiceNode := getServiceNodeName(cli, values.Service_Game)
		if gameServiceNode != "" {
			callErr = core.Rpc.AsyncCallWithoutResp(gameServiceNode, &game.CPlayerOffline{PlayerId: pid}, meta)
			if callErr != nil {
				mlog.Errorf("player %d call game.PlayerOffline failed, %v", pid, callErr)
			}
		}
	}
}
