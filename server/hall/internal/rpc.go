package internal

import (
	"context"
	"errors"
	"runtime/debug"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/values"
	"github.com/fixkme/othello/server/hall/internal/logic"
)

func RpcHandler(rc *rpc.RpcContext) {
	ctx := prepareContext(rc)
	argMsg, logicHandler := rc.ReqMsg, rc.Handler

	fn := func() {
		defer func() {
			if err := recover(); err != nil {
				mlog.Errorf("rpc handler panic: %v\n%s", err, debug.Stack())
				rc.ReplyErr = errors.New("rpc handler exception")
			}
			rc.SerializeResponse()
		}()

		rc.Reply, rc.ReplyErr = logicHandler(ctx, argMsg)

		if rc.ReplyErr == nil {
			mlog.Infof("handler msg succeed, method:%s, req_data:%v, rsp_data:%v", rc.MethodName, argMsg, rc.Reply)
		} else {
			mlog.Errorf("handler msg failed, method:%s, req_data:%v, err:%v", rc.MethodName, argMsg, rc.ReplyErr)
		}
	}
	//mlog.Debugf("game push handler method:%s", rc.Req.MethodName)
	if err := logic.GetGlobalLogic().TryRunFunc(fn); err != nil {
		mlog.Errorf("rpc handler push logic func failed: %v", err)
		rc.ReplyErr = err
		rc.SerializeResponse()
	}
}

func prepareContext(rc *rpc.RpcContext) (ctx context.Context) {
	ctx = context.WithValue(context.Background(), values.RpcContext, rc)
	ctx = context.WithValue(ctx, values.RpcContext_Meta, rc.ReqMd)
	return
}
