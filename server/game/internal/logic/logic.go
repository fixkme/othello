package logic

import (
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/gokit/util/app"
	mrpc "github.com/fixkme/othello/server/game/internal/rpc"
	"github.com/fixkme/othello/server/game/internal/services"
	"github.com/fixkme/othello/server/pb/game"
)

type LogicModule struct {
	name string
}

func NewLogicModule() app.Module {
	return &LogicModule{
		name: "logic",
	}
}

func (m *LogicModule) OnInit() error {
	err := mrpc.Module.GetRpcImp().RegisterService("game1", func(rpcSrv *rpc.Server, nodeName string) error {
		mlog.Info("RegisterService succeed %v", nodeName)
		game.RegisterGameServer(rpcSrv, &services.Service{})
		return nil
	})
	return err
}

func (m *LogicModule) Run() {

}

func (m *LogicModule) OnDestroy() {

}

func (m *LogicModule) Name() string {
	return m.name
}
