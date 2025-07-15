package shared

import (
	"context"
	"strings"
	"time"

	"github.com/fixkme/gokit/rpc"
	"google.golang.org/protobuf/proto"
)

// 同步调用
func SyncCall(ctx context.Context, cc *rpc.ClientConn, req proto.Message, outRsp proto.Message) (err error) {
	opt := &rpc.CallOption{
		Timeout: 3 * time.Second,
	}
	fullName := string(req.ProtoReflect().Descriptor().FullName())
	v2 := strings.SplitN(fullName, ".", 2)
	service, method := v2[0], v2[1][1:]
	_, _, err = cc.Invoke(ctx, service, method, req, outRsp, opt)
	return
}

// 异步调用，不需要回应
func AsyncCall(ctx context.Context, cc *rpc.ClientConn, req proto.Message) (err error) {
	opt := &rpc.CallOption{
		Async: true,
	}
	fullName := string(req.ProtoReflect().Descriptor().FullName())
	v2 := strings.SplitN(fullName, ".", 2)
	service, method := v2[0], v2[1][1:]
	_, _, err = cc.Invoke(ctx, service, method, req, nil, opt)
	return
}

// 异步调用，带有回应
func AsyncCallWithResp(ctx context.Context, cc *rpc.ClientConn, req proto.Message, outRsp proto.Message, outRet chan *rpc.AsyncCallResult, passData any) (err error) {
	opt := &rpc.CallOption{
		Async:        true,
		Timeout:      3 * time.Second,
		AsyncRetChan: outRet,
		PassThrough:  passData,
	}
	fullName := string(req.ProtoReflect().Descriptor().FullName())
	v2 := strings.SplitN(fullName, ".", 2)
	service, method := v2[0], v2[1][1:]
	_, _, err = cc.Invoke(ctx, service, method, req, outRsp, opt)
	return
}
