package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/util"
	"github.com/fixkme/othello/server/common/const/env"
	"github.com/fixkme/othello/server/common/framework"
	"github.com/fixkme/othello/server/common/shared"
	"github.com/fixkme/othello/server/game/internal"
)

func main() {
	start()
}

func start() {
	fmt.Println("start game server")

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := mlog.UseDefaultLogger(ctx, wg, "./logs", "game", "debug", true); err != nil {
		fmt.Println("UseDefaultLogger err:", err)
		return
	}

	if err := shared.InitMongo(env.GetEnvStr(env.APP_MongoUrl)); err != nil {
		fmt.Println("InitMongo err:", err)
		return
	}

	rpcModule := framework.CreateRpcModule("game_rpc", internal.DispatcherFunc, internal.RpcHandler)
	internal.RpcModule = rpcModule
	util.DefaultApp().Run(
		rpcModule,
		internal.NewLogicModule(),
	)

	cancel()
	wg.Wait()
}
