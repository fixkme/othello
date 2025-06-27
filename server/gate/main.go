package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/fixkme/gokit/log"
	"github.com/fixkme/gokit/util/app"
	"github.com/fixkme/othello/server/common/framework"
	"github.com/fixkme/othello/server/gate/internal"
)

func main() {
	start()
}

func start() {
	fmt.Println("start gate server")

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	if err := log.UseDefaultLogger(ctx, wg, "../logs", "gate", "debug", true); err != nil {
		panic(err)
	}

	rpcModule := framework.CreateRpcModule("game_rpc", nil, nil)
	internal.RpcModule = rpcModule
	app.DefaultApp().Run(
		internal.NewGateModule(),
		rpcModule,
	)

	cancel()
	wg.Wait()
}
