package logic

import "github.com/fixkme/gokit/util/app"

type LogicModule struct {
	name string
}

func NewLogicModule() app.Module {
	return &LogicModule{
		name: "logic",
	}
}

func (m *LogicModule) OnInit() error {
	return nil
}

func (m *LogicModule) Run() {

}

func (m *LogicModule) OnDestroy() {

}

func (m *LogicModule) Name() string {
	return m.name
}
