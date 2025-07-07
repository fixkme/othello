package internal

import (
	"context"
	"math/rand"

	"github.com/cloudwego/netpoll"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/const/values"
	"github.com/fixkme/othello/server/common/framework"
	"google.golang.org/protobuf/proto"
)

var RpcModule *framework.RpcModule
var RpcNodeName string

func DispatcherFunc(conn netpoll.Connection, rpcReq *rpc.RpcRequestMessage) int {
	md := rpcReq.GetMd()
	if md != nil {
		if v := md.GetInt(values.Rpc_SessionId); v != 0 {
			return int(v)
		}
	}
	return rand.Int()
}

func RpcHandlerFunc(rc *rpc.RpcContext, ser rpc.ServerSerializer) {
	argMsg, logicHandler := rc.Method(rc.SrvImpl)
	if err := proto.Unmarshal(rc.Req.Payload, argMsg); err == nil {
		rc.Reply, rc.ReplyErr = logicHandler(context.Background(), argMsg)
	} else {
		rc.ReplyErr = err
	}
	if rc.ReplyErr == nil {
		mlog.Info("gate handler msg succeed, req_data:%v, rsp_data:%v", argMsg, rc.Reply)
	} else {
		mlog.Error("gate handler msg failed, req_data:%v, err:%v", argMsg, rc.ReplyErr)
	}
	ser(rc, false)
}
