package system

import (
	"context"
	"fmt"
	"sync"
	"time"

	g "github.com/fixkme/gokit/framework/go"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/othello/server/game/internal/entity"
	"github.com/fixkme/othello/server/game/internal/logic"
	"github.com/fixkme/othello/server/game/internal/time_event"
)

type globalSystem struct {
	defaultModule
	logic *g.RoutineAgent

	playerInRooms      map[int64]int64 // 玩家所在房间
	roomAgentWaitGroup sync.WaitGroup
	roomAgents         map[int64]*entity.RoomAgent
	closingRooms       map[int64]chan struct{}

	timerGenId     int64
	timerCallbacks map[string]globalTimerCallback
	runningTimers  map[int64]*time_event.Global
	callbackTimers map[int64]int64

	rootCtx    context.Context
	rootCancel context.CancelFunc
}

type globalTimerCallback func(data time_event.Event, now int64)

// globalSystem 全局系统
var Global = new(globalSystem)

func init() {
	Manager.register(Global)
}

// onInit 初始化
func (s *globalSystem) onInit() {
	s.timerCallbacks = make(map[string]globalTimerCallback)
	s.runningTimers = make(map[int64]*time_event.Global)
	s.callbackTimers = make(map[int64]int64)
	s.roomAgents = make(map[int64]*entity.RoomAgent, 128)
	s.closingRooms = make(map[int64]chan struct{})
	s.playerInRooms = make(map[int64]int64)
	s.rootCtx, s.rootCancel = context.WithCancel(context.Background())
}

// afterInit 初始化后
func (s *globalSystem) afterInit() {
	s.logic = logic.GetGlobalLogic()

	go s.runTicker(s.rootCtx)
	// 注册定时器回调

	// 其他全局数据的加载

	// 其他全局模块的初始化

	// 完全初始化后再执行
	//s.createMidnightTimer()
}

// Close 关闭
func (s *globalSystem) Close() {
	mlog.Infof("global system closing")
	s.rootCancel()
}

// onClose 关闭
func (s *globalSystem) onClose() {

}

// generateTimerId 生成定时器id
// func (s *globalSystem) generateTimerId() int64 {

// }

func (s *globalSystem) SyncGetTargetRoom(ctx context.Context, roomId int64) (result *entity.RoomAgent, err error) {
	var closingSignal <-chan struct{}
	err = s.SyncExec(func() {
		if result = s.getRoomAgent(roomId); result != nil {
			closingSignal = s.closingRooms[roomId]
		}
	})

	if err != nil {
		return
	} else if result != nil {
		if closingSignal == nil {
			return
		}
		select {
		case <-closingSignal:
		case <-ctx.Done():
			err = ctx.Err()
			return
		}
	} else {
		err = fmt.Errorf("not find room %d", roomId)
		return
	}

	return
}

func (s *globalSystem) SyncGetTargetRoomByPlayer(ctx context.Context, playerId int64) (result *entity.RoomAgent, err error) {
	var ok bool
	var roomId int64
	var closingSignal <-chan struct{}
	err = s.SyncCall(func() error {
		roomId, ok = s.playerInRooms[playerId]
		if !ok {
			return fmt.Errorf("player  not exist in room pid %d", playerId)
		}
		if result = s.getRoomAgent(roomId); result != nil {
			closingSignal = s.closingRooms[roomId]
		}
		return nil
	})

	if err != nil {
		return
	} else if result != nil {
		if closingSignal == nil {
			return
		}
		select {
		case <-closingSignal:
			return nil, fmt.Errorf("room %d is closing", roomId)
		case <-ctx.Done():
			err = ctx.Err()
			return
		}
	} else {
		err = fmt.Errorf("player not find pid %d", playerId)
		return
	}
}

// SyncCall 同步调用
func (s *globalSystem) SyncCall(fn func() error) (err error) {
	fail := s.logic.SyncRunFunc(func() {
		err = fn()
	})
	if fail != nil {
		return fail
	}
	return err
}

// SyncExec 同步执行
func (s *globalSystem) SyncExec(fn func()) error {
	return s.logic.SyncRunFunc(fn)
}

// AsyncExec 异步执行
func (s *globalSystem) AsyncExec(fn func()) error {
	return s.logic.TryRunFunc(fn)
}

// MetricsCollect 收集监控指标
func (s *globalSystem) MetricsCollect() {
}

// runTicker 运行定时器
func (s *globalSystem) runTicker(ctx context.Context) {
	// 监控统计
	const metricsDuration = 5 * time.Second
	metricsTicker := time.NewTicker(metricsDuration)
	defer metricsTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-metricsTicker.C:
			s.AsyncExec(s.MetricsCollect)
		}
	}
}
