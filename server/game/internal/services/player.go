package services

import (
	"context"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/othello/server/pb/game"
)

func (s *Service) Login(ctx context.Context, in *game.CLogin) (*game.SLogin, error) {
	mlog.Info("game handler CLogin:%v", in)
	resp := &game.SLogin{
		ServerTz: 28800000,
	}
	return resp, nil
}

func (s *Service) EnterGame(ctx context.Context, in *game.CEnterGame) (*game.SEnterGame, error) {

	resp := &game.SEnterGame{}
	return resp, nil
}

func (s *Service) PlacePiece(ctx context.Context, in *game.CPlacePiece) (*game.SPlacePiece, error) {

	resp := &game.SPlacePiece{}
	return resp, nil
}
