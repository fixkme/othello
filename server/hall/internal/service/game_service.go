package service

import (
	"context"

	"github.com/fixkme/othello/server/hall/internal/system"
	"github.com/fixkme/othello/server/pb/hall"
)

func (s *Service) GameSettle(ctx context.Context, in *hall.CGameSettle) (*hall.SGameSettle, error) {
	system.Global.GameSettle(in)
	return &hall.SGameSettle{}, nil
}
