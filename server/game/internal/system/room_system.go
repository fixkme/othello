package system

import (
	"context"
	"errors"
	"time"

	"github.com/fixkme/gokit/clock"
	"github.com/fixkme/gokit/framework/core"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/util"
	"github.com/fixkme/othello/server/common/values"
	"github.com/fixkme/othello/server/game/internal/entity"
	"github.com/fixkme/othello/server/game/internal/time_event"
	"github.com/fixkme/othello/server/pb/datas"
	"github.com/fixkme/othello/server/pb/game"
	"github.com/fixkme/othello/server/pb/hall"
	"google.golang.org/protobuf/proto"
)

type roomSystem struct {
	defaultModule
	timerCallbacks map[string]roomTimerCallback
}

type roomTimerCallback func(room *entity.Room, data time_event.Event, now int64)

var Room = new(roomSystem)

func init() {
	Manager.register(Room)
}

func (s *roomSystem) onInit() {
	s.timerCallbacks = make(map[string]roomTimerCallback)
}

func (s *roomSystem) afterInit() {
	s.registerTimerCallback(&time_event.PlayerOffline{}, s.onPlayerOffline)
}

// 在nats协程执行
func (s *roomSystem) InitRoom(r *entity.Room, initPlayers []*datas.PBPlayerInfo, gateIds map[int64]string) (err error) {
	r.Desk.Init(initPlayers)
	r.Players[r.Desk.OwnerPlayer.Id] = r.Desk.OwnerPlayer
	r.Desk.OwnerPlayer.GateId = gateIds[r.Desk.OwnerPlayer.Id]
	if r.Desk.OppoPlayer != nil {
		r.Players[r.Desk.OppoPlayer.Id] = r.Desk.OppoPlayer
		r.Desk.OppoPlayer.GateId = gateIds[r.Desk.OppoPlayer.Id]
	}
	for _, p := range r.Players {
		Player.OnEnterRoom(p, r)
	}
	return
}

func (s *roomSystem) ReadyToStart(r *entity.Room) error {
	r.Desk.Status = datas.TableStatus_TS_WaitReady
	msg := &game.PReadyGame{
		TableInfo: r.Desk.PackPB(),
	}
	s.BroadcastPlayer(r, msg)
	return nil
}

// 广播房间内所有玩家消息
func (s *roomSystem) BroadcastPlayer(r *entity.Room, msg proto.Message, excepts ...int64) error {
	if len(r.Players) == 0 {
		return nil
	}
	var except int64
	if len(excepts) > 0 {
		except = excepts[0]
	}
	v := make([]*entity.Player, 0, len(r.Players))
	for _, p := range r.Players {
		if except == p.Id {
			continue
		}
		v = append(v, p)
	}
	if err := NoticePlayer(msg, v...); err != nil {
		mlog.Errorf("broadcastPlayer %v, error: %v", util.GetMessageFullName(msg), err)
		return err
	}
	return nil
}

func (s *roomSystem) OnEnd(r *entity.Room) {
	mlog.Infof("%d room retire, now players num %d", r.Desk.Id, len(r.Players))
	for _, p := range r.Players {
		if p.IsRobot {
			continue
		}
		Player.OnLeaveRoom(p.Id, r.Desk.Id)
	}
}

// SyncCall 同步调用
func (s *roomSystem) SyncCall(ctx context.Context, roomId int64, fn func(*entity.Room) error) (err error) {
	var agent *entity.RoomAgent
	if agent, err = Global.SyncGetTargetRoom(ctx, roomId); err != nil {
		return
	}
	var fail error
	err = agent.SyncRunFunc(func() {
		fail = fn(agent.Room)
	})
	if fail != nil {
		err = fail
	}
	return
}

func (s *roomSystem) SyncExec(ctx context.Context, roomId int64, fn func(*entity.Room)) (err error) {
	var agent *entity.RoomAgent
	if agent, err = Global.SyncGetTargetRoom(ctx, roomId); err != nil {
		return
	}
	return agent.SyncRunFunc(func() {
		fn(agent.Room)
	})
}

// AsyncExec 异步执行
func (s *roomSystem) AsyncExec(roomId int64, fn func(*entity.Room)) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		agent, err := Global.SyncGetTargetRoom(ctx, roomId)
		if err != nil {
			mlog.Errorf("get target room %d failed, %s", roomId, err)
			return
		}
		agent.MustRunFunc(func() {
			fn(agent.Room)
		})
	}()
}

func (s *roomSystem) PlayerReadyGame(r *entity.Room, p *entity.Player) error {
	r.Readys[p.Id] = true
	if len(r.Readys) == 2 {
		r.Desk.Status = datas.TableStatus_TS_Playing
		msg := &game.PStartGame{}
		s.BroadcastPlayer(r, msg)
	}
	return nil
}

func (s *roomSystem) PlayerLeaveGame(r *entity.Room, p *entity.Player) error {
	s.GameOver(r, p.GetPlayPieceType(), true)
	return nil
}

func (s *roomSystem) PlacePiece(r *entity.Room, p *entity.Player, x, y int, t entity.PieceType) error {
	if p.GetPlayPieceType() != t {
		return errors.New("player piece type not match")
	}
	tb := r.Desk
	if !tb.IsOperator(t) {
		return errors.New("operator not match")
	}
	if !tb.CanPlacePiece(x, y, t) {
		return errors.New("location cant reverse piece")
	}

	// 放下棋子
	//tb.Record()
	tb.AddPiece(x, y, t)
	tb.Reverse(x, y, t)
	// 对方能否有棋可翻转
	if tb.CanPlace(-t) {
		tb.TurnOperator()
		mlog.Debugf("change operator to %d", tb.Operator)
	} else {
		mlog.Debugf("%d no piece can reverse", -t)
	}
	// 广播落子结果
	msg := &game.PPlacePiece{
		PieceType: int32(t), X: int32(x), Y: int32(y), OperatePiece: int32(tb.Operator),
	}
	NoticePlayer(msg, tb.OwnerPlayer, tb.OppoPlayer)
	// 检查游戏是否达到结束条件
	if tb.CheckEnd() {
		var loser_piece_type entity.PieceType
		if tb.BlackCount > tb.WhiteCount {
			loser_piece_type = entity.PieceType_White
		} else if tb.BlackCount < tb.WhiteCount {
			loser_piece_type = entity.PieceType_Black
		} else {
			loser_piece_type = entity.PieceType_None
		}
		s.GameOver(r, loser_piece_type, false)
	}

	return nil
}

func (s *roomSystem) GameOver(r *entity.Room, loser_piece_type entity.PieceType, isGiveUp bool) {
	if r.Desk.Status == datas.TableStatus_TS_Over {
		return
	}
	r.Desk.Status = datas.TableStatus_TS_Over
	msg := &game.PGameResult{
		WinnerPieceType: -int64(loser_piece_type),
		LoserPieceType:  int64(loser_piece_type),
		IsGiveUp:        isGiveUp,
	}
	s.BroadcastPlayer(r, msg)

	tid := r.Desk.Id
	tb := r.Desk
	hallMsg := &hall.CGameSettle{
		OwnerPlayer: tb.OppoPlayer.PBPlayerInfo,
		OppoPlayer:  tb.OwnerPlayer.PBPlayerInfo,
		TableId:     tid,
	}
	core.Rpc.AsyncCallWithoutResp(values.Service_Hall, hallMsg)
	// 删除
	p1 := tb.OwnerPlayer.Id
	p2 := tb.OppoPlayer.Id
	s.removePlayer(r, tb.OwnerPlayer)
	s.removePlayer(r, tb.OppoPlayer)
	go func() {
		Global.logic.MustRunFunc(func() {
			Global.removePlayerFromRoom(p1, tid)
			Global.removePlayerFromRoom(p2, tid)
			Global.closeRoomAgent(tid)
		})
	}()
}

func (s *roomSystem) removePlayer(r *entity.Room, p *entity.Player) {
	delete(r.Players, p.Id)
	if p.OfflineTimer != 0 {
		clock.CancelTimer(p.OfflineTimer)
		p.OfflineTimer = 0
	}
}

func (s *roomSystem) PlayerOffline(r *entity.Room, p *entity.Player) {
	delay := int64(10 * 1e3)
	timerId, err := r.CreateTimer(delay, &time_event.PlayerOffline{PlayerId: p.Id})
	if err != nil {
		mlog.Errorf("create PlayerOffline timer failed,pid:%d, %v", p.Id, err)
		return
	}
	p.OfflineTimer = timerId
}

func (s *roomSystem) onPlayerOffline(r *entity.Room, ev time_event.Event, now int64) {
	data := ev.(*time_event.PlayerOffline)
	p := r.Players[data.PlayerId]
	if p == nil {
		return
	}
	s.PlayerLeaveGame(r, p)
}
