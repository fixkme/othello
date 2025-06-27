package rpc

import (
	"github.com/fixkme/othello/server/common/framework"
)

var Module *framework.RpcModule

// func DispatcherFunc(conn netpoll.Connection, rpcReq *rpc.RpcRequestMessage) int {
// 	md := rpcReq.GetMd()
// 	if md != nil {
// 		if v := md.GetInt(values.Rpc_SessionId); v != 0 {
// 			return int(v)
// 		}
// 	}
// 	return rand.Int()
// }

// func RpcHandler(rc *rpc.RpcContext, ser rpc.ServerSerializer) {
// 	argMsg, logicHandler := rc.Method(rc.SrvImpl)
// 	if err := proto.Unmarshal(rc.Req.Payload, argMsg); err == nil {
// 		rc.Reply, rc.ReplyErr = logicHandler(context.Background(), argMsg)
// 	} else {
// 		rc.ReplyErr = err
// 	}
// 	if rc.ReplyErr == nil {
// 		log.Info("game handler msg succeed, req_data:%v, rsp_data:%v", argMsg, rc.Reply)
// 	} else {
// 		log.Error("game handler msg failed, req_data:%v, err:%v", argMsg, rc.ReplyErr)
// 	}
// 	ser(rc, false)
// }
