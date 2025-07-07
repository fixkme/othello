package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/util/app"
	"github.com/fixkme/othello/server/common/framework"
	"github.com/fixkme/othello/server/game/internal/logic"
	"github.com/fixkme/othello/server/game/internal/rpc"
)

func main() {
	start()
}

// 启动etcd， rpcServer
// 开启定时器
// logic 准备数据，加载excel配置、加载mongo数据
// rpcServer注册service逻辑接口
// 向etcd注册ip、port，可以接受rpc请求

// 向etcd删除自己的ip、port
// stop logic， 保存数据
// stop 定时器
// stop rpcServer、etcd

func start() {
	fmt.Println("start gate server")

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	if err := mlog.UseDefaultLogger(ctx, wg, "./logs", "game", "debug", true); err != nil {
		panic(err)
	}

	rpcModule := framework.CreateRpcModule("game_rpc", nil, nil)
	rpc.Module = rpcModule
	app.DefaultApp().Run(
		rpcModule,
		logic.NewLogicModule(),
	)

	cancel()
	wg.Wait()
}
