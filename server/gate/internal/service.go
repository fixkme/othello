package internal

import (
	"context"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/othello/server/pb/gate"
)

type Service struct {
}

func (s *Service) NoticePlayer(ctx context.Context, in *gate.CNoticePlayer) (*gate.SNoticePlayer, error) {
	mlog.Info("handler CNoticePlayer %v", in)
	cli := ClientMgr.GetClient(in.PlayerId)
	if cli == nil {
		return nil, nil
	}
	// todo: send notice
	return nil, nil
}
