package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/values"
	"github.com/fixkme/othello/server/hall/internal/entity"
	"github.com/fixkme/othello/server/hall/internal/system"
)

type Service struct {
}

func getRpcContext(ctx context.Context) (*rpc.RpcContext, error) {
	rc, ok := ctx.Value(values.RpcContext).(*rpc.RpcContext)
	if !ok {
		return nil, errors.New("context does not contain rpc.RpcContext")
	}
	return rc, nil
}
func getMetaWithContext(ctx context.Context) (val *rpc.Meta, err error) {
	md, ok := ctx.Value(values.RpcContext_Meta).(*rpc.Meta)
	if !ok {
		return nil, errors.New("context does not contain rpc.Meta")
	}
	return md, nil
}

func getPlayerWithContext(ctx context.Context) (player *entity.Player, err error) {
	md, err := getMetaWithContext(ctx)
	if err != nil {
		return
	}
	playerId := md.GetInt(values.Rpc_PlayerId)
	if playerId == 0 {
		return nil, errors.New("context not exist playerId")
	}
	player = system.Global.GetPlayer(playerId)
	if player == nil {
		return nil, fmt.Errorf("players not exist %d", playerId)
	}
	if v := md.GetStr(values.Rpc_GateId); v != "" {
		player.SetGateId(v)
	} else {
		mlog.Warnf("rpc context meta not exist gateId, %d", playerId)
	}
	return
}
