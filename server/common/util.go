package common

import (
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/values"
)

func WarpMeta(playerid int64, gateId string) *rpc.Meta {
	md := &rpc.Meta{}
	if playerid > 0 {
		md.SetInt(values.Rpc_PlayerId, playerid)
	}
	if gateId != "" {
		md.SetStr(values.Rpc_GateId, gateId)
	}
	return md
}
