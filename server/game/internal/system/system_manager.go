package system

import (
	"reflect"
	"runtime/debug"

	"github.com/fixkme/gokit/mlog"
)

type systemModule interface {
	onInit()
	afterInit()
	onClose()
}

type defaultModule struct {
}

func (m *defaultModule) onInit() {
}

func (m *defaultModule) afterInit() {
}

func (m *defaultModule) onClose() {
}

type systemManager struct {
	exists  map[string]struct{}
	systems []systemModule
}

var Manager = new(systemManager)

func (m *systemManager) register(s systemModule) {
	m.systems = append(m.systems, s)
	if m.exists == nil {
		m.exists = make(map[string]struct{})
	}
	name := m.parseName(s)
	if _, ok := m.exists[name]; ok {
		panic("system has been registered: " + name)
	}
	m.exists[name] = struct{}{}
}

func (m *systemManager) parseName(s systemModule) string {
	t := reflect.TypeOf(s)
	return t.String()
}

func (m *systemManager) Init() {
	for _, s := range m.systems {
		s.onInit()
	}
}

func (m *systemManager) AfterInit() {
	defer func() {
		if r := recover(); r != nil {
			callStack := string(debug.Stack())
			mlog.Fatalf("after init panic: %v\n%s", r, callStack)
		}
	}()

	for _, s := range m.systems {
		s.afterInit()
	}
}

func (m *systemManager) Close() {
	for _, s := range m.systems {
		s.onClose()
	}
}
