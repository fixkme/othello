package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/const/values"
	"github.com/fixkme/othello/server/pb/game"
)

type Service struct {
}

func getMdContextValue(ctx context.Context) (val *rpc.Meta, err error) {
	md, ok := ctx.Value(RpcMdContext).(*rpc.Meta)
	if !ok {
		return nil, errors.New("context does not contain rpc.Meta")
	}
	return md, nil
}

func getRpcContextValue(ctx context.Context) (val *rpc.RpcContext, err error) {
	rc, ok := ctx.Value(RpcContext).(*rpc.RpcContext)
	if !ok {
		return nil, errors.New("context does not contain rpc.RpcContext")
	}
	return rc, nil
}

func getPlayer(ctx context.Context) (player *Player, err error) {
	md, err := getMdContextValue(ctx)
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
	player.GateId = md.GetStr(values.Rpc_GateId)
	return
}

func (s *Service) Login(ctx context.Context, in *game.CLogin) (*game.SLogin, error) {
	mlog.Debug("game handler CLogin:%v", in)
	md, err := getMdContextValue(ctx)
	if err != nil {
		return nil, err
	}
	rc, err := getRpcContextValue(ctx)
	if err != nil {
		return nil, err
	}

	acct := in.Account
	if acct == "" {
		return nil, errors.New("account is empty")
	}
	p := global.GetPlayerByAccount(acct)
	if p == nil {
		p = global.CreatePlayer(acct)
	}
	var inTable int64
	if p.PlayingTable != nil {
		inTable = p.PlayingTable.Id
	}
	// 记录gateId
	p.GateId = md.GetStr(values.Rpc_GateId)
	// 设置session id
	replyMd := &rpc.Meta{}
	replyMd.SetInt(values.Rpc_SessionId, p.Id())
	rc.ReplyMd = replyMd

	resp := &game.SLogin{
		PlayerData: p.ToPB(),
		ServerTz:   28800000,
		TableId:    inTable,
	}
	return resp, nil
}

func (s *Service) EnterGame(ctx context.Context, in *game.CEnterGame) (*game.SEnterGame, error) {
	p, err := getPlayer(ctx)
	if err != nil {
		return nil, err
	}
	resp := &game.SEnterGame{}
	if tb := p.PlayingTable; tb != nil {
		// 恢复
		resp.TableInfo = tb.PackPB()
	} else {
		// 创建或匹配
		tb = global.CheckMatchTable(p)
		if tb == nil {
			tb = global.CreateTable(p)
			global.AddMatching(tb)
		}
		resp.TableInfo = tb.PackPB()
	}

	return resp, nil
}

func (s *Service) LeaveGame(ctx context.Context, in *game.CLeaveGame) (*game.SLeaveGame, error) {
	p, err := getPlayer(ctx)
	if err != nil {
		return nil, err
	}
	global.PlayerLeaveGame(p)
	return &game.SLeaveGame{}, nil
}

func (s *Service) PlayerOffline(ctx context.Context, in *game.CPlayerOffline) (*game.SPlayerOffline, error) {
	p := global.GetPlayer(in.PlayerId)
	if p != nil {
		global.PlayerLeaveGame(p)
		global.RemovePlayer(p.Id())
	}
	return nil, nil
}

func (s *Service) PlacePiece(ctx context.Context, in *game.CPlacePiece) (*game.SPlacePiece, error) {
	p, err := getPlayer(ctx)
	if err != nil {
		return nil, err
	}
	tb := p.PlayingTable
	if tb == nil {
		return nil, errors.New("player not in playing table")
	}
	err = tb.PlacePiece(int(in.X), int(in.Y), PieceType(in.PieceType))
	resp := &game.SPlacePiece{}
	return resp, err
}
