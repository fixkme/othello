package internal

import (
	"reflect"
	"strings"

	"github.com/fixkme/gokit/mlog"
)

func (g *Global) registerTimerCallback(event any, callback func(data any, now int64)) {
	name := GetEventName(event)
	g.timerCallbacks[name] = callback
}

func (g *Global) onTimerTrigger(event any, now int64) {
	name := GetEventName(event)
	cb, ok := g.timerCallbacks[name]
	if !ok {
		mlog.Error("timer callback not register %v", name)
		return
	}
	cb(event, now)
}

func GetEventName(event any) string {
	name := reflect.TypeOf(event).String()
	if index := strings.LastIndexAny(name, "."); index != -1 {
		name = name[index+1:]
	}
	return name
}
