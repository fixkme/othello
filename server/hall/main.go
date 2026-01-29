package main

import (
	"context"
	"sync"

	"github.com/fixkme/gokit/framework/app"
	"github.com/fixkme/gokit/framework/core"
	"github.com/fixkme/othello/server/common"
	"github.com/fixkme/othello/server/common/values"
	"github.com/fixkme/othello/server/hall/internal"
)

func main() {
	run()
}

func run() {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	common.StartApp(ctx, wg, values.Service_Hall, internal.RpcHandler)

	app.DefaultApp().Run(
		core.Rpc,
		internal.NewLogicModule(),
	)

	cancel()
	wg.Wait()
}
