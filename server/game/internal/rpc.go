package internal

import (
	"context"
	"errors"
	"runtime/debug"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/values"
	"github.com/fixkme/othello/server/game/internal/entity"
	"github.com/fixkme/othello/server/game/internal/system"
)

func RpcHandler(rc *rpc.RpcContext) {
	ctx, agent := prepareContext(rc)
	argMsg, logicHandler := rc.ReqMsg, rc.Handler

	fn := func() {
		defer func() {
			if err := recover(); err != nil {
				mlog.Errorf("game rpc handler panic: %v\n%s", err, debug.Stack())
				rc.ReplyErr = errors.New("rpc handler exception")
			}
			rc.SerializeResponse()
		}()

		rc.Reply, rc.ReplyErr = logicHandler(ctx, argMsg)

		if rc.ReplyErr == nil {
			mlog.Infof("game handler msg succeed, method:%s, req_data:{%v}, rsp_data:{%v}", rc.MethodName, argMsg, rc.Reply)
		} else {
			mlog.Errorf("game handler msg failed, method:%s, req_data:{%v}, err:%v", rc.MethodName, argMsg, rc.ReplyErr)
		}
	}
	//mlog.Debugf("game push handler method:%s", rc.Req.MethodName)
	if agent != nil {
		if err := agent.TryRunFunc(fn); err != nil {
			mlog.Errorf("rpc handler push agent func failed, %v", err)
			rc.ReplyErr = errors.New("rpc handler push agent func failed")
			rc.SerializeResponse()
		}
		return
	}
	if err := system.Global.AsyncExec(fn); err != nil {
		mlog.Errorf("rpc handler push logic func failed, err:%v", err)
		rc.ReplyErr = errors.New("rpc handler push logic func failed")
		rc.SerializeResponse()
		return
	}
}

func prepareContext(rc *rpc.RpcContext) (ctx context.Context, agent *entity.RoomAgent) {
	ctx = context.WithValue(context.Background(), values.RpcContext, rc)
	if md := rc.ReqMd; md != nil {
		ctx = context.WithValue(ctx, values.RpcContext_Meta, md)
		if playerId := md.GetInt(values.Rpc_PlayerId); playerId != 0 {
			agent, _ = system.Global.SyncGetTargetRoomByPlayer(ctx, playerId)
			if agent != nil {
				ctx = context.WithValue(ctx, values.RpcContext_RoomAgent, agent)
			}
		}
	}
	return
}
