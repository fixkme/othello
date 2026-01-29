package system

import (
	"fmt"
	"runtime/debug"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/othello/server/game/internal/entity"
	"github.com/fixkme/othello/server/game/internal/time_event"
)

func (s *roomSystem) registerTimerCallback(ev time_event.Event, cb roomTimerCallback) {
	name := time_event.GetEventName(ev)
	if _, exists := s.timerCallbacks[name]; exists {
		panic(fmt.Sprintf("time_event callback %s has been registered", name))
	}
	s.timerCallbacks[name] = cb
}

func (s *roomSystem) OnTimerCallback(r *entity.Room, t *time_event.Desk, now int64) {
	//mlog.Debugf("roomSystem.OnTimerCallback rid:%d, now:%d, tid:%d, data:%#v", r.Id, now, t.CallbackId, t.Event)
	defer func() {
		if err := recover(); err != nil {
			mlog.Errorf("roomSystem.OnTimerCallback panic: %v, stack:%v", err, debug.Stack())
		}
	}()
	name := time_event.GetEventName(t.Event)
	cb, ok := s.timerCallbacks[name]
	if !ok {
		mlog.Errorf("%d timer callback not register %v", r.Desk.Id, name)
		return
	} else if t.EndTime > now {
		mlog.Errorf("%d timer callback not deadline %v", r.Desk.Id, name)
		return
	}
	cb(r, t.Event, now)
}
