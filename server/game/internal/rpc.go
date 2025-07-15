package internal

import (
	"context"
	"errors"
	"math/rand"

	"github.com/cloudwego/netpoll"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/const/values"
	"github.com/fixkme/othello/server/common/framework"
	"google.golang.org/protobuf/proto"
)

var RpcModule *framework.RpcModule

type LogicContextKeyType string

const LogicContextKey LogicContextKeyType = "logicContext"

// type LogicContext struct {
// 	Md *rpc.Meta
// }

func DispatcherFunc(conn netpoll.Connection, rpcReq *rpc.RpcRequestMessage) int {
	md := rpcReq.GetMd()
	if md != nil {
		if v := md.GetInt(values.Rpc_SessionId); v != 0 {
			return int(v)
		}
	}
	return rand.Int()
}

func RpcHandler(rc *rpc.RpcContext, ser rpc.ServerSerializer) {
	ctx := prepareContext(rc)
	argMsg, logicHandler := rc.Method(rc.SrvImpl)
	if err := proto.Unmarshal(rc.Req.Payload, argMsg); err != nil {
		rc.ReplyErr = err
		ser(rc, false)
	}

	fn := func() {
		defer func() {
			if err := recover(); err != nil {
				mlog.Error("game rpc handler panic: %v", err)
				rc.ReplyErr = errors.New("rpc handler exception")
			}
			ser(rc, false)
		}()

		rc.Reply, rc.ReplyErr = logicHandler(ctx, argMsg)

		if rc.ReplyErr == nil {
			mlog.Info("game handler msg succeed, req_data:%v, rsp_data:%v", argMsg, rc.Reply)
		} else {
			mlog.Error("game handler msg failed, req_data:%v, err:%v", argMsg, rc.ReplyErr)
		}
	}

	if err := logicModule.PushLogicFunc(fn); err != nil {
		mlog.Error("game rpc handler push logic func failed: %v", err)
		rc.ReplyErr = err
		ser(rc, false)
	}
}

func prepareContext(rc *rpc.RpcContext) (ctx context.Context) {
	ctx = context.WithValue(context.Background(), LogicContextKey, rc.Req.Md)
	return
}
