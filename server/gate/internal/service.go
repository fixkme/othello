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
		mlog.Debugf("NoticePlayer notices is empty")
		return
	}
	playerId := in.PlayerId
	cli := ClientMgr.GetClient(in.PlayerId)
	if cli == nil {
		mlog.Debugf("NoticePlayer not exist player %v", playerId)
		return
	}
	msg := &ws.WsResponseMessage{Notices: in.Notices}
	content, err := proto.Marshal(msg)
	if err != nil {
		mlog.Errorf("NoticePlayer marshal err:%v", err)
		return
	}
	err = cli.conn.Send(content)
	if err != nil {
		mlog.Errorf("NoticePlayer %d err:%v", playerId, err)
	}
	return
}

func (s *Service) BroadcastPlayer(ctx context.Context, in *gate.CBroadcastPlayer) (_ *gate.SBroadcastPlayer, _ error) {
	if len(in.PlayerIds) == 0 {
		mlog.Debugf("BroadcastPlayer playerIds is empty")
		return
	}
	if len(in.Notices) == 0 {
		mlog.Debugf("BroadcastPlayer notices is empty")
		return
	}

	msg := &ws.WsResponseMessage{Notices: in.Notices}
	content, err := proto.Marshal(msg)
	if err != nil {
		mlog.Errorf("BroadcastPlayer marshal err:%v", err)
		return
	}

	for _, playerId := range in.PlayerIds {
		cli := ClientMgr.GetClient(playerId)
		if cli == nil {
			mlog.Debugf("BroadcastPlayer not exist player %v", playerId)
			continue
		}
		err = cli.conn.Send(content)
		if err != nil {
			mlog.Errorf("BroadcastPlayer %d err:%v", playerId, err)
		}
	}
	return
}
