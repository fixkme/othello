package services

import (
	"context"

	"github.com/fixkme/othello/server/pb/game"
)

func (s *Service) Login(ctx context.Context, in *game.CLogin) (*game.SLogin, error) {

	resp := &game.SLogin{}
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
