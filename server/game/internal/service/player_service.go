package service

import (
	"context"
	"errors"

	"github.com/fixkme/gokit/clock"
	"github.com/fixkme/othello/server/game/internal/entity"
	"github.com/fixkme/othello/server/game/internal/system"
	"github.com/fixkme/othello/server/pb/game"
)

func (s *Service) PlayerOffline(ctx context.Context, in *game.CPlayerOffline) (*game.SPlayerOffline, error) {
	p, r, err := getPlayerRoomWithContext(ctx)
	if err != nil {
		return nil, err
	}
	system.Room.PlayerOffline(r, p)
	return nil, nil
}

// 重进房间恢复
func (s *Service) EnterGame(ctx context.Context, in *game.CEnterGame) (*game.SEnterGame, error) {
	resp := &game.SEnterGame{}
	p, r, err := getPlayerRoomWithContext(ctx)
	if err != nil {
		return nil, err
	}
	if p.OfflineTimer > 0 {
		clock.CancelTimer(p.OfflineTimer)
		p.OfflineTimer = 0
	}
	resp.TableInfo = r.Desk.PackPB()
	return resp, nil
}

func (s *Service) LeaveGame(ctx context.Context, in *game.CLeaveGame) (*game.SLeaveGame, error) {
	p, r, err := getPlayerRoomWithContext(ctx)
	if err != nil {
		return nil, err
	}
	if err = system.Room.PlayerLeaveGame(r, p); err != nil {
		return nil, err
	}
	return &game.SLeaveGame{}, nil
}

func (s *Service) ReadyGame(ctx context.Context, in *game.CReadyGame) (*game.SReadyGame, error) {
	p, r, err := getPlayerRoomWithContext(ctx)
	if err != nil {
		return nil, err
	}
	if err = system.Room.PlayerReadyGame(r, p); err != nil {
		return nil, err
	}
	return &game.SReadyGame{}, nil
}

func (s *Service) PlacePiece(ctx context.Context, in *game.CPlacePiece) (*game.SPlacePiece, error) {
	p, r, err := getPlayerRoomWithContext(ctx)
	if err != nil {
		return nil, err
	}
	tb := r.Desk
	if tb == nil {
		return nil, errors.New("player not in playing table")
	}
	err = system.Room.PlacePiece(r, p, int(in.X), int(in.Y), entity.PieceType(in.PieceType))
	resp := &game.SPlacePiece{}
	return resp, err
}
