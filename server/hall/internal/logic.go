package internal

import (
	"github.com/fixkme/gokit/clock"
	"github.com/fixkme/gokit/framework/app"
	"github.com/fixkme/gokit/framework/core"
	g "github.com/fixkme/gokit/framework/go"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/values"
	"github.com/fixkme/othello/server/hall/internal/logic"
	"github.com/fixkme/othello/server/hall/internal/service"
	"github.com/fixkme/othello/server/hall/internal/system"
	"github.com/fixkme/othello/server/hall/internal/time_event"
	"github.com/fixkme/othello/server/pb/hall"
)

type LogicModule struct {
	*g.RoutineAgent
	quit chan struct{}
	name string
}

var logicModule *LogicModule

func NewLogicModule() app.Module {
	logicModule = &LogicModule{
		RoutineAgent: g.NewRoutineAgent(1024, 1024),
		quit:         make(chan struct{}, 1),
		name:         "hall_logic",
	}
	return logicModule
}

func (m *LogicModule) OnInit() error {
	err := core.Rpc.RegisterServiceOnlyOne(values.Service_Hall, func(rpcSrv rpc.ServiceRegistrar, nodeName string) error {
		mlog.Infof("RegisterService succeed %v", nodeName)
		hall.RegisterHallServer(rpcSrv, &service.Service{})
		return nil
	})
	if err != nil {
		return err
	}
	clock.Start(m.quit)
	m.RoutineAgent.Init(m.onTimerCallback, m.beforeClose)
	logic.InitLogic(m.RoutineAgent)
	system.Manager.Init()
	m.MustSubmit(m.afterInit)
	return nil
}

func (m *LogicModule) afterInit() {
	system.Manager.AfterInit()
}

func (m *LogicModule) beforeClose() {
	mlog.Infof("logic module before close")
	system.Global.Close()
	system.Manager.Close()
}

func (m *LogicModule) Destroy() {
	if err := core.Rpc.UnregisterService(values.Service_Hall); err != nil {
		mlog.Errorf("Rpc.UnregisterService failed, %v", err)
	}
	m.RoutineAgent.Close()
}

func (m *LogicModule) Name() string {
	return m.name
}

func (m *LogicModule) onTimerCallback(callbackId int64, now int64, data any) {
	switch d := data.(type) {
	case *time_event.Global:
		system.Global.OnTimerCallback(callbackId, now, d)
	default:
		mlog.Errorf("undefined timer event type %T", data)
	}
}
