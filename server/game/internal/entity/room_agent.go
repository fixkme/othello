package entity

import (
	g "github.com/fixkme/gokit/framework/go"
)

type RoomAgent struct {
	*Room
	*g.RoutineAgent
}

func NewRoomAgent(desk *Room, tcb g.TimerCb) *RoomAgent {
	a := &RoomAgent{
		Room:         desk,
		RoutineAgent: g.NewRoutineAgent(1024, 1024),
	}
	a.RoutineAgent.Init(tcb, a.onClose)
	a.Room.timerReceiver = a.RoutineAgent.GetTimerReciver()
	return a
}

func (a *RoomAgent) onClose() {
	a.Room.roomClosed = true
}
