package internal

import (
	"context"
	"errors"
	"runtime/debug"

	"github.com/fixkme/gokit/framework/core"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/values"
	"google.golang.org/protobuf/proto"
)

func RpcHandler(rc *rpc.RpcContext) {
	ctx := prepareContext(rc)
	argMsg, logicHandler := rc.Method(rc.SrvImpl)
	if err := proto.Unmarshal(rc.Req.Payload, argMsg); err != nil {
		rc.ReplyErr = err
		rc.SerializeResponse(nil)
		return
	}

	fn := func() {
		defer func() {
			if err := recover(); err != nil {
				mlog.Errorf("game rpc handler panic: %v\n%s", err, debug.Stack())
				rc.ReplyErr = errors.New("rpc handler exception")
			}
			rc.SerializeResponse(&core.MsgMarshaler)
		}()

		rc.Reply, rc.ReplyErr = logicHandler(ctx, argMsg)

		if rc.ReplyErr == nil {
			mlog.Infof("game handler msg succeed, method:%s, req_data:%v, rsp_data:%v", rc.Req.MethodName, argMsg, rc.Reply)
		} else {
			mlog.Errorf("game handler msg failed, method:%s, req_data:%v, err:%v", rc.Req.MethodName, argMsg, rc.ReplyErr)
		}
	}
	//mlog.Debugf("game push handler method:%s", rc.Req.MethodName)
	if err := logicModule.PushLogicFunc(fn); err != nil {
		mlog.Errorf("game rpc handler push logic func failed: %v", err)
		rc.ReplyErr = err
		rc.SerializeResponse(nil)
	}
}

func prepareContext(rc *rpc.RpcContext) (ctx context.Context) {
	ctx = context.WithValue(context.Background(), values.RpcContext, rc)
	ctx = context.WithValue(ctx, values.RpcContext_Meta, rc.Req.Md)
	return
}
