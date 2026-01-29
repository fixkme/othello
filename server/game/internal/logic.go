package internal

import (
	"errors"
	"fmt"

	"github.com/fixkme/gokit/clock"
	"github.com/fixkme/gokit/framework/app"
	"github.com/fixkme/gokit/framework/config"
	"github.com/fixkme/gokit/framework/core"
	g "github.com/fixkme/gokit/framework/go"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/values"
	"github.com/fixkme/othello/server/game/internal/logic"
	"github.com/fixkme/othello/server/game/internal/service"
	"github.com/fixkme/othello/server/game/internal/system"
	"github.com/fixkme/othello/server/game/internal/time_event"
	"github.com/fixkme/othello/server/pb/game"
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
		name:         "game_logic",
	}
	return logicModule
}

func (m *LogicModule) OnInit() error {
	if config.Config.ServerId == 0 {
		return errors.New("config server_id not set")
	}

	serviceNodeName := fmt.Sprintf("%s.%d", values.Service_Game, config.Config.ServerId)
	err := core.Rpc.RegisterServiceOnlyOne(serviceNodeName, func(rpcSrv rpc.ServiceRegistrar, nodeName string) error {
		mlog.Infof("RegisterService succeed %v", nodeName)
		game.RegisterGameServer(rpcSrv, &service.Service{})
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
	system.Global.WaitCloseAllRoomAgents()
	system.Global.Close()
	system.Manager.Close()
}

func (m *LogicModule) Destroy() {
	serviceNodeName := fmt.Sprintf("%s.%d", values.Service_Game, config.Config.ServerId)
	if err := core.Rpc.UnregisterService(serviceNodeName); err != nil {
		mlog.Errorf("Rpc.UnregisterService failed, %s, %v", serviceNodeName, err)
	}
	m.RoutineAgent.Close()
	close(m.quit)
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
