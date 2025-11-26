package internal

import (
	"context"
	"errors"
	"math/rand"
	"runtime/debug"

	"github.com/cloudwego/netpoll"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/const/values"
	"github.com/fixkme/othello/server/common/framework"
	"google.golang.org/protobuf/proto"
)

var RpcModule *framework.RpcModule

type LogicContextKeyType string

const (
	RpcContext   LogicContextKeyType = "RpcContext"
	RpcMdContext LogicContextKeyType = "RpcMdContext"
)

func DispatcherFunc(conn netpoll.Connection, rpcReq *rpc.RpcRequestMessage) int {
	md := rpcReq.GetMd()
	if md != nil {
		if v := md.GetInt(values.Rpc_SessionId); v != 0 {
			return int(v)
		}
	}
	return rand.Int()
}

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
				mlog.Error("game rpc handler panic: %v\n%s", err, debug.Stack())
				rc.ReplyErr = errors.New("rpc handler exception")
			}
			rc.SerializeResponse(&framework.Marshaler)
		}()

		rc.Reply, rc.ReplyErr = logicHandler(ctx, argMsg)

		if rc.ReplyErr == nil {
			mlog.Info("game handler msg succeed, method:%s, req_data:%v, rsp_data:%v", rc.Req.MethodName, argMsg, rc.Reply)
		} else {
			mlog.Error("game handler msg failed, method:%s, req_data:%v, err:%v", rc.Req.MethodName, argMsg, rc.ReplyErr)
		}
	}
	//mlog.Debug("game push handler method:%s", rc.Req.MethodName)
	if err := logicModule.PushLogicFunc(fn); err != nil {
		mlog.Error("game rpc handler push logic func failed: %v", err)
		rc.ReplyErr = err
		rc.SerializeResponse(nil)
	}
}

func prepareContext(rc *rpc.RpcContext) (ctx context.Context) {
	ctx = context.WithValue(context.Background(), RpcContext, rc)
	ctx = context.WithValue(ctx, RpcMdContext, rc.Req.Md)
	return
}
