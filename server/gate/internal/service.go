package internal

import (
	"context"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/othello/server/pb/gate"
	"github.com/fixkme/othello/server/pb/ws"
	"google.golang.org/protobuf/proto"
)

type Service struct {
}

func (s *Service) NoticePlayer(ctx context.Context, in *gate.CNoticePlayer) (_ *gate.SNoticePlayer, _ error) {
	if len(in.Notices) == 0 {
		mlog.Debug("NoticePlayer notices is empty")
		return
	}
	playerId := in.PlayerId
	cli := ClientMgr.GetClient(in.PlayerId)
	if cli == nil {
		mlog.Debug("NoticePlayer not exist player %v", playerId)
		return
	}
	msg := &ws.WsResponseMessage{Notices: in.Notices}
	content, err := proto.Marshal(msg)
	if err != nil {
		mlog.Error("NoticePlayer marshal err:%v", err)
		return
	}
	err = cli.conn.Send(content)
	if err != nil {
		mlog.Error("NoticePlayer %d err:%v", playerId, err)
	}
	return
}

func (s *Service) BroadcastPlayer(ctx context.Context, in *gate.CBroadcastPlayer) (_ *gate.SBroadcastPlayer, _ error) {
	if len(in.PlayerIds) == 0 {
		mlog.Debug("BroadcastPlayer playerIds is empty")
		return
	}
	if len(in.Notices) == 0 {
		mlog.Debug("BroadcastPlayer notices is empty")
		return
	}

	msg := &ws.WsResponseMessage{Notices: in.Notices}
	content, err := proto.Marshal(msg)
	if err != nil {
		mlog.Error("BroadcastPlayer marshal err:%v", err)
		return
	}

	for _, playerId := range in.PlayerIds {
		cli := ClientMgr.GetClient(playerId)
		if cli == nil {
			mlog.Debug("BroadcastPlayer not exist player %v", playerId)
			continue
		}
		err = cli.conn.Send(content)
		if err != nil {
			mlog.Error("BroadcastPlayer %d err:%v", playerId, err)
		}
	}
	return
}
