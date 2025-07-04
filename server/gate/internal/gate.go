package internal

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/util/app"
	"github.com/fixkme/gokit/wsg"
	"github.com/fixkme/othello/server/common/const/env"
	"github.com/panjf2000/gnet/v2"
)

type GateServer struct {
	wsServer   *wsg.Server
	routerPool *_LoadBalanceImp
	name       string
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
	}
	listenAddr := env.GetEnvStr(env.APP_GateListenAddr)
	opt := &wsg.ServerOptions{
		Options:       gnetOpt,
		Addr:          fmt.Sprintf("tcp://%s", listenAddr), // "tcp://:7070",
		OnHandshake:   func(conn *wsg.Conn, r *http.Request) error { return s.routerPool.OnHandshake(conn, r) },
		OnClientClose: OnClientClose,
		OnServerShutdown: func(_ gnet.Engine) {
			s.routerPool.Stop()
			mlog.Info("ws server shutdown")
		},
	}
	s.wsServer = wsg.NewServer(opt)
	s.routerPool = newRouterPool(4, 1024)
	return nil
}

func (s *GateServer) Run() {
	s.routerPool.Start()
	if err := s.wsServer.Run(); err != nil {
		mlog.Error("wsServer Run error: %v", err)
	}
	mlog.Info("ws server exit run")
}

func (s *GateServer) OnDestroy() {
	if err := s.wsServer.Stop(context.Background()); err != nil {
		mlog.Error("wsServer Stop error:%v", err)
	}
}

func (s *GateServer) Name() string {
	return s.name
}
