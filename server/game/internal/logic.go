package internal

import (
	"fmt"

	"github.com/fixkme/gokit/clock"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/gokit/util"
	"github.com/fixkme/othello/server/common/const/values"
	"github.com/fixkme/othello/server/common/framework"
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

func NewLogicModule() util.Module {
	logicModule = &LogicModule{
		fnChan:      make(chan func(), 1024),
		timerCbChan: make(chan *clock.Promise, 1024),
		quit:        make(chan struct{}, 1),
		name:        "game_logic",
	}
	return logicModule
}

func (m *LogicModule) OnInit() error {
	clock.Start(m.quit)
	m.timerCaller = global.onTimerTrigger
	serviceNodeName := fmt.Sprintf("%s.%d", values.Service_Game, 1)
	err := framework.Rpc.RegisterService(serviceNodeName, func(rpcSrv rpc.ServiceRegistrar, nodeName string) error {
		mlog.Info("RegisterService succeed %v", nodeName)
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
			mlog.Debug("game LogicModule quit")
			return
		case fn := <-m.fnChan:
			fn()
		case promise := <-m.timerCbChan:
			m.timerCaller(promise.Data, promise.NowTs)
		}
	}
}

func (m *LogicModule) OnDestroy() {
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
