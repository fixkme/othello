package service

import (
	"context"

	"github.com/fixkme/othello/server/game/internal/system"
	"github.com/fixkme/othello/server/pb/game"
)

func (s *Service) CreateRoom(ctx context.Context, in *game.CCreateRoom) (*game.SCreateRoom, error) {
	resp := &game.SCreateRoom{}
	roomId := in.TableId
	err := system.Global.OpenNewRoom(roomId, in.Players, in.GateIds)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
