package framework

import (
	"context"
	"math/rand"
	"time"

	"github.com/cloudwego/netpoll"
	"github.com/fixkme/gokit/log"
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
		ListenAddr:     listenAddr,
		PollerNum:      4,
		ProcessorSize:  7,
		DispatcherFunc: dispatcher,
		HandlerFunc:    handlerFunc,
	}
	etcdOpt := &etcd.EtcdOpt{
		Config: clientv3.Config{
			Endpoints:            []string{"127.0.0.1:2379"},
			DialTimeout:          5 * time.Second,
			DialKeepAliveTime:    5 * time.Second,
			DialKeepAliveTimeout: 3 * time.Second,
			AutoSyncInterval:     5 * time.Second,
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
		log.Error("%v module stop error: %v", m.name, err)
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

// 默认RpcHandler
func RpcHandlerFunc(rc *rpc.RpcContext, ser rpc.ServerSerializer) {
	argMsg, logicHandler := rc.Method(rc.SrvImpl)
	if err := proto.Unmarshal(rc.Req.Payload, argMsg); err == nil {
		rc.Reply, rc.ReplyErr = logicHandler(context.Background(), argMsg)
	} else {
		rc.ReplyErr = err
	}
	if rc.ReplyErr == nil {
		log.Info("rpc handler msg succeed, req_data:%v, rsp_data:%v", argMsg, rc.Reply)
	} else {
		log.Error("rpc handler msg failed, req_data:%v, err:%v", argMsg, rc.ReplyErr)
	}
	ser(rc, false)
}
