package entity

import (
	"errors"

	"github.com/fixkme/gokit/clock"
	"github.com/fixkme/gokit/util"
	"github.com/fixkme/othello/server/game/internal/time_event"
)

type Room struct {
	Desk    *Table
	Players map[int64]*Player
	Readys  map[int64]bool

	roomClosed    bool
	timerReceiver chan<- *clock.Promise
}

func NewRoom(tid int64) *Room {
	tb := &Room{
		Desk:    NewTable(tid),
		Players: make(map[int64]*Player),
		Readys:  make(map[int64]bool),
	}
	return tb
}

func (r *Room) CreateTimer(delay int64, event time_event.Event) (timerId int64, err error) {
	if r.roomClosed {
		return 0, errors.New("room closed")
	}
	endTime := util.NowMs() + delay
	t := &time_event.Desk{
		Event:   event,
		EndTime: endTime,
	}
	t.CallbackId, err = clock.NewTimer(endTime, t, r.timerReceiver, nil)
	if err != nil {
		return
	}
	return t.CallbackId, nil
}
