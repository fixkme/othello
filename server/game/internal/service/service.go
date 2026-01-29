package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"

	"github.com/fixkme/othello/server/common/values"
	"github.com/fixkme/othello/server/game/internal/entity"
	"github.com/fixkme/othello/server/pb/datas"
)

type Service struct {
}

func getRpcContext(ctx context.Context) (val *rpc.RpcContext, err error) {
	rc, ok := ctx.Value(values.RpcContext).(*rpc.RpcContext)
	if !ok {
		return nil, errors.New("context does not contain rpc.RpcContext")
	}
	return rc, nil
}

func GetMetaWithContext(ctx context.Context) (val *rpc.Meta, err error) {
	md, ok := ctx.Value(values.RpcContext_Meta).(*rpc.Meta)
	if !ok {
		return nil, errors.New("context does not contain rpc.Meta")
	}
	return md, nil
}

func getPlayerRoomWithContext(ctx context.Context) (player *entity.Player, room *entity.Room, err error) {
	md, err := GetMetaWithContext(ctx)
	if err != nil {
		return
	}
	playerId := md.GetInt(values.Rpc_PlayerId)
	if playerId == 0 {
		return nil, nil, errors.New("rpc meta is not set playerId")
	}

	agent, ok := ctx.Value(values.RpcContext_RoomAgent).(*entity.RoomAgent)
	if !ok {
		err = errors.New("context does not contain rpc.RoomAgent")
		return
	}
	room = agent.Room
	player, ok = room.Players[playerId]
	if !ok {
		err = fmt.Errorf("player %v not in room %d", playerId, room.Desk.Id)
		return
	}
	if room.Desk.Status == datas.TableStatus_TS_Over {
		err = fmt.Errorf("game is over")
		return
	}
	if gateId := md.GetStr(values.Rpc_GateId); gateId != "" {
		player.GateId = gateId
	} else {
		mlog.Warnf("rpc meta does not contain gateId of player %d", playerId)
	}
	return
}
