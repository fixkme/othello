package logic

import g "github.com/fixkme/gokit/framework/go"

var globalLogic *g.RoutineAgent

func InitLogic(l *g.RoutineAgent) {
	globalLogic = l
}

func GetGlobalLogic() *g.RoutineAgent {
	return globalLogic
}
