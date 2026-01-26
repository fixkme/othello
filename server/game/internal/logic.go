package internal

import (
	"errors"
	"fmt"

	"github.com/fixkme/gokit/clock"
	"github.com/fixkme/gokit/framework/app"
	"github.com/fixkme/gokit/framework/config"
	"github.com/fixkme/gokit/framework/core"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/values"
	"github.com/fixkme/othello/server/pb/game"
)

type LogicModule struct {
	fnChan      chan func()
	timerCbChan chan *clock.Promise
	timerCaller func(data any, now int64)
	quit        chan struct{}
	name        string
}

var logicModule *LogicModule

func NewLogicModule() app.Module {
	logicModule = &LogicModule{
		fnChan:      make(chan func(), 1024),
		timerCbChan: make(chan *clock.Promise, 1024),
		quit:        make(chan struct{}, 1),
		name:        "game_logic",
	}
	return logicModule
}

func (m *LogicModule) OnInit() error {
	if config.Config.ServerId == 0 {
		return errors.New("config server_id not set")
	}
	clock.Start(m.quit)
	m.timerCaller = global.onTimerTrigger
	serviceNodeName := fmt.Sprintf("%s.%d", values.Service_Game, config.Config.ServerId)
	err := core.Rpc.RegisterServiceOnlyOne(serviceNodeName, func(rpcSrv rpc.ServiceRegistrar, nodeName string) error {
		mlog.Infof("RegisterService succeed %v", nodeName)
		game.RegisterGameServer(rpcSrv, &Service{})
		return nil
	})
	if err := global.Init(); err != nil {
		return fmt.Errorf("init global failed err:%v", err)
	}
	return err
}

func (m *LogicModule) Run() {
	defer func() {
		global.OnRetire()
	}()

	for {
		select {
		case <-m.quit:
			mlog.Debugf("game LogicModule quit")
			return
		case fn := <-m.fnChan:
			fn()
		case promise := <-m.timerCbChan:
			m.timerCaller(promise.Data, promise.NowTs)
		}
	}
}

func (m *LogicModule) Destroy() {
	serviceNodeName := fmt.Sprintf("%s.%d", values.Service_Game, config.Config.ServerId)
	if err := core.Rpc.UnregisterService(serviceNodeName); err != nil {
		mlog.Errorf("Rpc.UnregisterService failed, %s, %v", serviceNodeName, err)
	}
	close(m.quit)
}

func (m *LogicModule) Name() string {
	return m.name
}

func (m *LogicModule) CreateTimer(when int64, data any) (timerId int64, err error) {
	return clock.NewTimer(when, data, m.timerCbChan, nil)
}

func (m *LogicModule) CancelTimer(timerId int64) {
	clock.CancelTimer(timerId)
}

func (m *LogicModule) PushLogicFunc(f func()) error {
	m.fnChan <- f
	return nil
}
