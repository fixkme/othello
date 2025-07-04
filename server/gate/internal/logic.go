package internal

import (
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/pb/gate"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/util/app"
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
	err := RpcModule.GetRpcImp().RegisterService("gate1", func(rpcSrv *rpc.Server, nodeName string) error {
		mlog.Info("RegisterService succeed %v", nodeName)
		gate.RegisterGateServer(rpcSrv, &Service{})
		return nil
	})
	if err != nil {
		mlog.Error("RegisterService failed %v", err)
		return err
	}
	return nil
}

func (m *LogicModule) Run() {

}

func (m *LogicModule) OnDestroy() {

}

func (m *LogicModule) Name() string {
	return m.name
}
