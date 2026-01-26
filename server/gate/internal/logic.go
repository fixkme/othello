package internal

import (
	"github.com/fixkme/gokit/framework/app"
	"github.com/fixkme/gokit/framework/core"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/values"
	"github.com/fixkme/othello/server/pb/gate"

	"github.com/fixkme/gokit/mlog"
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
	err := core.Rpc.RegisterServiceOnlyOne(values.Service_Gate, func(rpcSrv rpc.ServiceRegistrar, nodeName string) error {
		mlog.Infof("RegisterService succeed %v", nodeName)
		GateNodeId = nodeName
		gate.RegisterGateServer(rpcSrv, &Service{})
		return nil
	})
	if err != nil {
		mlog.Errorf("RegisterService failed %v", err)
		return err
	}
	return err
}

func (m *LogicModule) Run() {

}

func (m *LogicModule) Destroy() {

}

func (m *LogicModule) Name() string {
	return m.name
}
