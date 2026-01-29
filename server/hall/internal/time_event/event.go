package time_event

import (
	"encoding/gob"
	"log"
	"reflect"
	"strings"
)

type Event interface {
	timerEvent()
}

type Default struct {
}

func (d *Default) timerEvent() {
}

var events = make(map[string]Event)

func register(event Event) {
	name := GetEventName(event)
	if _, ok := events[name]; ok {
		log.Fatalf("event %s already registered", name)
	}
	events[name] = event
	gob.Register(event)
}

func GetEventName(event Event) string {
	name := reflect.TypeOf(event).String()
	if index := strings.LastIndexAny(name, "."); index != -1 {
		name = name[index+1:]
	}
	return name
}
