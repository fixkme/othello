package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/const/values"
	"github.com/fixkme/othello/server/game/internal/entity"
	"github.com/fixkme/othello/server/pb/game"
)

type Service struct {
}

func getContextValue(ctx context.Context) (val *rpc.Meta, err error) {
	md, ok := ctx.Value(LogicContextKey).(*rpc.Meta)
	if !ok {
		return nil, errors.New("context does not contain rpc.Meta")
	}
	return md, nil
}

func getPlayer(ctx context.Context) (player *entity.Player, err error) {
	md, err := getContextValue(ctx)
	if err != nil {
		return
	}
	playerId := md.GetInt(values.Rpc_SessionId)
	if playerId == 0 {
		return nil, errors.New("context not exist playerId")
	}
	player = global.GetPlayer(playerId)
	if player == nil {
		return nil, fmt.Errorf("players not exist %d", playerId)
	}
	return
}

func (s *Service) Login(ctx context.Context, in *game.CLogin) (*game.SLogin, error) {
	mlog.Debug("game handler CLogin:%v", in)
	player := global.CreatePlayer()
	resp := &game.SLogin{
		PlayerData: player.ToPB(),
		ServerTz:   28800000,
	}
	return resp, nil
}

func (s *Service) EnterGame(ctx context.Context, in *game.CEnterGame) (*game.SEnterGame, error) {
	player, err := getPlayer(ctx)
	if err != nil {
		return nil, err
	}
	resp := &game.SEnterGame{}
	if player.PlayingTable != nil {
		// 恢复
		resp.TableInfo = player.PlayingTable.PBTableInfo
	} else {
		// 创建或匹配

	}

	return resp, nil
}

func (s *Service) PlacePiece(ctx context.Context, in *game.CPlacePiece) (*game.SPlacePiece, error) {

	resp := &game.SPlacePiece{}
	return resp, nil
}

func (s *Service) PlayerOffline(ctx context.Context, in *game.CPlayerOffline) (*game.SPlayerOffline, error) {
	global.RemovePlayer(in.PlayerId)
	return nil, nil
}
