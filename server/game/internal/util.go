package internal

import (
	"context"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/shared"
	"github.com/fixkme/othello/server/pb/gate"
	"github.com/fixkme/othello/server/pb/ws"
	"google.golang.org/protobuf/proto"
)

func NoticePlayer(msg proto.Message, players ...*Player) error {
	if len(players) == 0 {
		return nil
	}
	msgType := string(msg.ProtoReflect().Descriptor().FullName())
	msgData, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	rpcImp := RpcModule.GetRpcImp()
	for _, p := range players {
		pmsg := &gate.CNoticePlayer{
			PlayerId: p.Id(),
			Notices:  []*ws.PBPackage{{MessageType: msgType, MessagePayload: msgData}},
		}
		_, err = rpcImp.Call(p.GateId, func(ctx context.Context, cc *rpc.ClientConn) (proto.Message, error) {
			if err := shared.AsyncCallWithoutResp(ctx, cc, pmsg); err != nil {
				return nil, err
			}
			return nil, nil
		})
		if err != nil {
			mlog.Error("NotifyPlayer %d error: %v", p.Id(), err)
		}
	}
	return nil
}
