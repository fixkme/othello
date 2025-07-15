package internal

import (
	"fmt"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/gokit/util/app"
	"github.com/fixkme/gokit/util/timer"
	"github.com/fixkme/othello/server/common/const/values"
	"github.com/fixkme/othello/server/pb/game"
)

type LogicModule struct {
	fnChan      chan func()
	timerCbChan chan *timer.Promise
	quit        chan struct{}
	name        string
}

var logicModule *LogicModule

func NewLogicModule() app.Module {
	logicModule = &LogicModule{
		fnChan:      make(chan func(), 1024),
		timerCbChan: make(chan *timer.Promise, 1024),
		quit:        make(chan struct{}, 1),
		name:        "game_logic",
	}
	return logicModule
}

func (m *LogicModule) OnInit() error {
	timer.Start(m.quit)
	serviceNodeName := fmt.Sprintf("%s.%d", values.Service_Game, 1)
	err := RpcModule.GetRpcImp().RegisterService(serviceNodeName, func(rpcSrv *rpc.Server, nodeName string) error {
		mlog.Info("RegisterService succeed %v", nodeName)
		game.RegisterGameServer(rpcSrv, &Service{})
		return nil
	})

	return err
}

func (m *LogicModule) Run() {
	for {
		select {
		case <-m.quit:
			return
		case fn := <-m.fnChan:
			fn()
		case promise := <-m.timerCbChan:
			m.onTimerTrigger(promise)
			return
		}
	}
}

func (m *LogicModule) OnDestroy() {
	close(m.quit)
}

func (m *LogicModule) Name() string {
	return m.name
}

func (m *LogicModule) onTimerTrigger(promise *timer.Promise) {
	mlog.Info("onTimerTrigger: timerId:%d, nowTs:%d, data:%v", promise.TimerId, promise.NowTs, promise.Data)

}

func (m *LogicModule) CreateTimer(when int64, data any) (timerId int64, err error) {
	return timer.NewTimer(when, data, m.timerCbChan, nil)
}

func (m *LogicModule) CancelTimer(timerId int64) {
	timer.CancelTimer(timerId)
}

func (m *LogicModule) PushLogicFunc(f func()) error {
	m.fnChan <- f
	return nil
}
