package internal

import (
	"context"
	"log"
	"net/http"

	"github.com/fixkme/gokit/util/app"
	"github.com/fixkme/gokit/wsg"
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
	opt := &wsg.ServerOptions{
		Options:     gnetOpt,
		Addr:        "tcp://:2333",
		OnHandshake: func(conn *wsg.Conn, r *http.Request) error { return s.routerPool.OnHandshake(conn, r) },
		OnServerShutdown: func(_ gnet.Engine) {
			s.routerPool.Stop()
			log.Println("ws server shutdown")
		},
	}
	s.wsServer = wsg.NewServer(opt)
	s.routerPool = newRouterPool(4, 1024)
	return nil
}

func (s *GateServer) Run() {
	s.routerPool.Start()
	if err := s.wsServer.Run(); err != nil {
		log.Fatalf("wsServer Run error: %v", err)
	}
	log.Println("ws server exit run")
}

func (s *GateServer) OnDestroy() {
	if err := s.wsServer.Stop(context.Background()); err != nil {
		log.Println("wsServer Stop error:", err)
	}
}

func (s *GateServer) Name() string {
	return s.name
}
