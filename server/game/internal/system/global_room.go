package system

import (
	"errors"
	"runtime/debug"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/othello/server/game/internal/entity"
	"github.com/fixkme/othello/server/game/internal/time_event"
	"github.com/fixkme/othello/server/pb/datas"
)

// 创建房间，并且启动协程
func (s *globalSystem) OpenNewRoom(roomId int64, initPlayers []*datas.PBPlayerInfo, gateIds map[int64]string) error {
	if len(initPlayers) == 0 {
		return errors.New("player empty")
	}

	room := entity.NewRoom(roomId)
	if err := Room.InitRoom(room, initPlayers, gateIds); err != nil {
		return err
	}
	if err := Room.ReadyToStart(room); err != nil {
		return err
	}

	err := s.SyncExec(func() {
		s.createRoomAgent(room)
		for _, p := range initPlayers {
			s.addPlayerToRoom(p.Id, roomId)
		}
	})
	if err != nil {
		mlog.Errorf("open room failed %d, %v", roomId, err)
	} else {
		mlog.Infof("open room success %d", roomId)
	}

	return err
}

func (s *globalSystem) WaitCloseAllRoomAgents() {
	for id := range s.roomAgents {
		s.closeRoomAgent(id)
	}
	s.roomAgentWaitGroup.Wait()
}

func (s *globalSystem) closeRoomAgent(roomId int64) {
	agent := s.getRoomAgent(roomId)
	if agent == nil {
		return
	}
	if _, exists := s.closingRooms[roomId]; exists {
		return
	}
	s.closingRooms[roomId] = make(chan struct{})
	go agent.Close()
}

func (s *globalSystem) getRoomAgent(id int64) *entity.RoomAgent {
	result := s.roomAgents[id]
	return result
}

func (s *globalSystem) roomAgentCount() int64 {
	return int64(len(s.roomAgents))
}

func (s *globalSystem) createRoomAgent(r *entity.Room) *entity.RoomAgent {
	agent := entity.NewRoomAgent(r, func(tid, now int64, data any) {
		ev := data.(*time_event.Desk)
		Room.OnTimerCallback(r, ev, now)
	})
	s.roomAgents[r.Desk.Id] = agent
	s.roomAgentWaitGroup.Add(1)
	go s.runRoomAgent(agent)
	return agent
}

func (s *globalSystem) runRoomAgent(a *entity.RoomAgent) {
	defer func() {
		roomId := a.Room.Desk.Id
		mlog.Infof("%d runRoomAgent end", roomId)
		if err := recover(); err != nil {
			mlog.Errorf("run room %d agent exception, %s, %s", roomId, err, debug.Stack())
		}
	}()
	defer s.roomAgentWaitGroup.Done()
	defer s.AsyncExec(func() {
		s.removeRoomAgent(a)
	})
	defer Room.OnEnd(a.Room)

	a.Run()
}

func (s *globalSystem) removeRoomAgent(a *entity.RoomAgent) {
	id := a.Room.Desk.Id
	delete(s.roomAgents, id)
	if signal := s.closingRooms[id]; signal != nil {
		close(signal)
		delete(s.closingRooms, id)
	}
}

func (s *globalSystem) addPlayerToRoom(pid, roomId int64) {
	if v, ok := s.playerInRooms[pid]; !ok {
		s.playerInRooms[pid] = roomId
	} else {
		mlog.Errorf("%d player already in room %d, current room %d", pid, roomId, v)
	}
}

func (s *globalSystem) removePlayerFromRoom(pid, roomId int64) {
	if v, ok := s.playerInRooms[pid]; ok && v == roomId {
		delete(s.playerInRooms, pid)
	} else {
		mlog.Errorf("%d player not in room %d, current room %d", pid, roomId, v)
	}
}

func (s *globalSystem) GetPlayerRoom(pid int64) (roomId int64, err error) {
	if v, ok := s.playerInRooms[pid]; ok {
		return v, nil
	}
	return 0, errors.New("player not in any room")
}
