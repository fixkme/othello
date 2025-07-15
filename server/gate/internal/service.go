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
	cli := ClientMgr.GetClient(in.PlayerId)
	if cli == nil {
		mlog.Debug("NoticePlayer not exist player %v", in.PlayerId)
		return
	}
	msg := &ws.WsPushMessage{Notices: in.Notices}
	content, err := proto.Marshal(msg)
	if err != nil {
		mlog.Error("NoticePlayer marshal err:%v", err)
		return
	}
	err = cli.conn.Send(content)
	if err != nil {
		mlog.Error("NoticePlayer %d err:%v", in.PlayerId, err)
	}
	return
}
