package framework

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/cloudwego/netpoll"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/gokit/servicediscovery/impl/etcd"
	"github.com/fixkme/othello/server/common/const/env"
	"github.com/fixkme/othello/server/common/const/values"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
)

type RpcModule struct {
	serverOpt *rpc.ServerOpt
	etcdOpt   *etcd.EtcdOpt
	rpcer     *rpc.RpcImp
	name      string
}

func CreateRpcModule(name string, dispatcher rpc.DispatchHash, handlerFunc rpc.RpcHandler) *RpcModule {
	if dispatcher == nil {
		dispatcher = DispatcherFunc
	}
	if handlerFunc == nil {
		handlerFunc = RpcHandlerFunc
	}
	listenAddr := env.GetEnvStr(env.APP_RpcListenAddr)
	serverOpt := &rpc.ServerOpt{
		ListenAddr:        listenAddr,
		PollerNum:         4,
		ProcessorSize:     7,
		ProcessorTaskSize: 1024,
		DispatcherFunc:    dispatcher,
		HandlerFunc:       handlerFunc,
	}
	etcdAddrs := env.GetEnvStr(env.APP_EtcdEndpoints)
	etcdOpt := &etcd.EtcdOpt{
		Config: clientv3.Config{
			Endpoints:            strings.Split(etcdAddrs, ","),
			DialTimeout:          5 * time.Second,
			DialKeepAliveTime:    5 * time.Second,
			DialKeepAliveTimeout: 3 * time.Second,
			AutoSyncInterval:     15 * time.Second,
		},
	}
	return &RpcModule{
		serverOpt: serverOpt,
		etcdOpt:   etcdOpt,
		name:      name,
	}
}

func (m *RpcModule) GetRpcImp() *rpc.RpcImp {
	return m.rpcer
}

func (m *RpcModule) OnInit() error {
	rpcAddr := env.GetEnvStr(env.APP_RpcAddr)
	rpcTmp, err := rpc.NewRpc(context.Background(), rpcAddr, "gbs", m.etcdOpt, m.serverOpt)
	if err != nil {
		return err
	}
	m.rpcer = rpcTmp
	return nil
}

func (m *RpcModule) Run() {
	if err := m.rpcer.Run(); err != nil {
		panic(err)
	}
}

func (m *RpcModule) OnDestroy() {
	err := m.rpcer.Stop()
	if err != nil {
		mlog.Error("%v module stop error: %v", m.name, err)
	}
}

func (m *RpcModule) Name() string {
	return m.name
}

// 默认Dispatcher
func DispatcherFunc(conn netpoll.Connection, rpcReq *rpc.RpcRequestMessage) int {
	md := rpcReq.GetMd()
	if md != nil {
		if v := md.GetInt(values.Rpc_SessionId); v != 0 {
			return int(v)
		}
	}
	return rand.Int()
}

var (
	Unmarshaler = proto.UnmarshalOptions{
		AllowPartial:   true, //跳过required字段检查，因为根本没有required字段
		DiscardUnknown: false,
		RecursionLimit: 100,
		NoLazyDecoding: true, //禁用lazy解码，1、没有lazy标志的字段，2、开启lazy字段解码会导致buff延长生命周期
	}
	Marshaler = proto.MarshalOptions{
		AllowPartial:  true,
		Deterministic: false, // 结果序列化不需要唯一（顺序保证）
	}
)

// 默认RpcHandler
func RpcHandlerFunc(rc *rpc.RpcContext) {
	argMsg, logicHandler := rc.Method(rc.SrvImpl)
	if err := Unmarshaler.Unmarshal(rc.Req.Payload, argMsg); err == nil {
		rc.Reply, rc.ReplyErr = logicHandler(context.Background(), argMsg)
	} else {
		rc.ReplyErr = err
	}
	if rc.ReplyErr == nil {
		mlog.Info("rpc handler msg succeed, method:%s, req_data:%v, rsp_data:%v", rc.Req.MethodName, argMsg, rc.Reply)
	} else {
		mlog.Error("rpc handler msg failed, method:%s, req_data:%v, err:%v", rc.Req.MethodName, argMsg, rc.ReplyErr)
	}
	rc.SerializeResponse(&Marshaler)
}
