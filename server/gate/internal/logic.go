package internal

import (
	"fmt"

	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/const/values"
	"github.com/fixkme/othello/server/common/framework"
	"github.com/fixkme/othello/server/pb/gate"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/util"
)

type LogicModule struct {
	name string
}

func NewLogicModule() util.Module {
	return &LogicModule{
		name: "logic",
	}
}

func (m *LogicModule) OnInit() error {
	serviceNodeName := fmt.Sprintf("%s.%d", values.Service_Gate, 1)
	err := framework.Rpc.RegisterService(serviceNodeName, func(rpcSrv rpc.ServiceRegistrar, nodeName string) error {
		mlog.Infof("RegisterService succeed %v", nodeName)
		framework.RpcNodeName = nodeName
		gate.RegisterGateServer(rpcSrv, &Service{})
		return nil
	})
	if err != nil {
		mlog.Errorf("RegisterService failed %v", err)
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
