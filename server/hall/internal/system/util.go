package system

import (
	"context"
	"fmt"

	"github.com/fixkme/gokit/framework/core"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/values"
	"github.com/fixkme/othello/server/hall/internal/entity"
	"github.com/fixkme/othello/server/pb/gate"
	"github.com/fixkme/othello/server/pb/ws"
	"google.golang.org/protobuf/proto"
)

func GameService(gameId int64) string {
	gameService := fmt.Sprintf("%s.%d", values.Service_Game, gameId)
	return gameService
}

func NoticePlayer(msg proto.Message, players ...*entity.Player) error {
	if len(players) == 0 {
		return nil
	}
	msgType := string(msg.ProtoReflect().Descriptor().FullName())
	msgData, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	for _, p := range players {
		pmsg := &gate.CNoticePlayer{
			PlayerId: p.Id(),
			Notices:  []*ws.PBPackage{{MessageType: msgType, MessagePayload: msgData}},
		}
		_, err = core.Rpc.Call(p.GateId, func(ctx context.Context, cc *rpc.ClientConn) (proto.Message, error) {
			if err := rpc.AsyncCallWithoutResp(ctx, cc, pmsg); err != nil {
				return nil, err
			}
			return nil, nil
		})
		if err != nil {
			mlog.Errorf("NotifyPlayer %d error: %v", p.Id(), err)
		}
	}
	return nil
}
