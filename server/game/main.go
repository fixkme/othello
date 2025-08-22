package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/util/app"
	"github.com/fixkme/othello/server/common/framework"
	"github.com/fixkme/othello/server/game/internal"
)

func main() {
	start()
}

func start() {
	fmt.Println("start game server")

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	if err := mlog.UseDefaultLogger(ctx, wg, "./logs", "game", "debug", true); err != nil {
		panic(err)
	}

	rpcModule := framework.CreateRpcModule("game_rpc", internal.DispatcherFunc, internal.RpcHandler)
	internal.RpcModule = rpcModule
	app.DefaultApp().Run(
		rpcModule,
		internal.NewLogicModule(),
	)

	cancel()
	wg.Wait()
}
