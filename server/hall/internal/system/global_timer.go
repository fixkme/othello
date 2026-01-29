package system

import (
	"fmt"

	"github.com/fixkme/gokit/clock"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/util"
	"github.com/fixkme/othello/server/hall/internal/time_event"
)

// registerTimerCallback 注册定时器回调
func (s *globalSystem) registerTimerCallback(e time_event.Event, callback globalTimerCallback) {
	name := time_event.GetEventName(e)
	if _, exists := s.timerCallbacks[name]; exists {
		panic(fmt.Sprintf("time_event callback %s has been registered", name))
	}
	s.timerCallbacks[name] = callback
}

// OnTimerCallback 定时器回调
func (s *globalSystem) OnTimerCallback(callbackId int64, now int64, t *time_event.Global) {
	timerId, exists := s.callbackTimers[callbackId]
	if !exists {
		// 可能在定时器回调前一刻刚好取消了
		return
	}

	delete(s.callbackTimers, callbackId)
	delete(s.runningTimers, timerId)

	name := time_event.GetEventName(t.Event)
	callback, exists := s.timerCallbacks[name]
	if !exists {
		return
	}

	mlog.Infof("timer callback %s, now %d, args %+v", name, now, t.Event)
	callback(t.Event, now)
}

// CreateTimer 创建定时器
func (s *globalSystem) CreateTimer(delay int64, event time_event.Event) (timerId int64, err error) {
	timerId = s.generateTimerId()

	// 从定时器模块创建一个定时器
	endTime := util.NowMs() + delay
	t := &time_event.Global{
		Event:   event,
		EndTime: endTime,
	}

	t.CallbackId, err = clock.NewTimer(endTime, t, s.logicRoutine.GetTimerReciver(), nil)
	if err != nil {
		return
	}

	// 将定时器id和数据绑定
	s.addTimer(timerId, t)

	return
}

// addTimer 添加定时器
func (s *globalSystem) addTimer(timerId int64, t *time_event.Global) {
	s.runningTimers[timerId] = t
	s.callbackTimers[t.CallbackId] = timerId
}

// UpdateTimer 更新定时器
func (s *globalSystem) UpdateTimer(timerId int64, endTime int64) (err error) {
	globalTimer := s.runningTimers[timerId]
	if globalTimer == nil {
		return
	}

	globalTimer.EndTime = endTime
	_, err = clock.UpdateTimer(globalTimer.CallbackId, endTime)
	if err != nil {
		return
	}
	return
}

// CancelTimer 取消定时器
func (s *globalSystem) CancelTimer(timerId int64) {
	globalTimer, exists := s.runningTimers[timerId]
	if !exists {
		return
	}

	delete(s.runningTimers, timerId)
	delete(s.callbackTimers, globalTimer.CallbackId)
	if _, err := clock.CancelTimer(globalTimer.CallbackId); err != nil {
		return
	}
}
