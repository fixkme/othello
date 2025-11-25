package internal

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/gokit/util"
	"github.com/fixkme/gokit/wsg"
	"github.com/fixkme/othello/server/common/const/env"
	"github.com/fixkme/othello/server/common/const/values"
	"github.com/fixkme/othello/server/common/shared"
	"github.com/fixkme/othello/server/pb/game"
	"github.com/panjf2000/gnet/v2"
	"google.golang.org/protobuf/proto"
)

type GateServer struct {
	wsServer   *wsg.Server
	routerPool *_LoadBalanceImp
	name       string
	retired    bool // server是否Shutdown
}

func NewGateModule() util.Module {
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
		Options:       gnetOpt,
		Addr:          fmt.Sprintf("tcp4://%s", listenAddr), // "tcp4://:7070",
		OnHandshake:   func(conn *wsg.Conn, r *http.Request) error { return s.routerPool.OnHandshake(conn, r) },
		OnClientClose: s.OnClientClose,
		OnServerShutdown: func(_ gnet.Engine) {
			s.routerPool.Stop()
			mlog.Info("gate server shutdowned")
		},
	}
	s.wsServer = wsg.NewServer(opt)
	s.routerPool = newRouterPool(4, 1024)
	mlog.Info("gate server listenAddr: (%s)", listenAddr)
	return nil
}

func (s *GateServer) Run() {
	s.routerPool.Start()
	if err := s.wsServer.Run(); err != nil {
		mlog.Error("GateServer Run error: %v", err)
		panic(err)
	}
	mlog.Info("gate server exit run")
}

func (s *GateServer) OnDestroy() {
	mlog.Info("gate server stop")
	s.retired = true
	if err := s.wsServer.Stop(context.Background()); err != nil {
		mlog.Error("wsServer Stop error:%v", err)
	}
	mlog.Info("gate server stoped")
}

func (s *GateServer) Name() string {
	return s.name
}

func (s *GateServer) OnClientClose(conn *wsg.Conn, err error) {
	if s.retired {
		// 主动关闭ws server
		return
	}
	cli := conn.GetSession().(*WsClient)
	pid := cli.PlayerId
	mlog.Info("player ws closed, acc:%s, pid:%d, addr:%s, reason:%v", cli.Account, pid, conn.RemoteAddr().String(), err)
	if pid > 0 {
		ClientMgr.RemoveClient(pid)
		// 通知game玩家下线
		gameServiceNode := getServiceNodeName(cli, values.Service_Game)
		_, callErr := RpcModule.GetRpcImp().Call(gameServiceNode, func(ctx context.Context, cc *rpc.ClientConn) (proto.Message, error) {
			_err := shared.AsyncCall(ctx, cc, &game.CPlayerOffline{PlayerId: pid})
			return nil, _err
		})
		if callErr != nil {
			mlog.Error("player %d call PlayerOffline failed, %v", pid, callErr)
		}
	}
}
