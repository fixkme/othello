package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/gokit/util/errs"
	"github.com/fixkme/gokit/wsg"
	"github.com/fixkme/othello/server/common/const/values"
	"github.com/fixkme/othello/server/pb/ws"
	"google.golang.org/protobuf/proto"
)

type RoutingTask struct {
	Cli   *WsClient
	Datas []byte
}

type RoutingWorkerImp struct {
	routingChan  chan *RoutingTask
	rpcReplyChan chan *rpc.AsyncCallResult
}

func (r *RoutingWorkerImp) RoutingMsg(task *RoutingTask) {
	client := task.Cli
	// 反序列化数据
	wsMessage := &ws.WsRequestMessage{}
	err := proto.Unmarshal(task.Datas, wsMessage)
	if err != nil {
		mlog.Error("proto.Unmarshal wsReq err:%v", err)
		client.conn.Close()
		return
	}
	// 检查是否登录
	if client.PlayerId == 0 && wsMessage.MsgName != "game.CLogin" {
		mlog.Error("first msg must game.CLogin")
		client.conn.Close()
		return
	}

	defer func() {
		if r := recover(); r != nil {
			mlog.Error("routing player %d panic: %v", client.PlayerId, r)
			err = errors.New("routing panic")
		}
		if err != nil {
			replyClientResponse(client, wsMessage.Uuid, "", nil, nil, err)
		}
	}()

	v2 := strings.SplitN(wsMessage.MsgName, ".", 2)
	if len(v2) != 2 {
		err = fmt.Errorf("msgName invalid")
		return
	}
	service, method := v2[0], v2[1][1:]
	serviceNode := getServiceNodeName(client, service)
	// 路由数据到game或其他
	RpcModule.GetRpcImp().Call(serviceNode, func(ctx context.Context, cc *rpc.ClientConn) (proto.Message, error) {
		md := &rpc.Meta{}
		md.AddStr(values.Rpc_GateId, RpcNodeName)
		if client.PlayerId != 0 {
			md.AddInt(values.Rpc_SessionId, client.PlayerId)
		}
		callOpt := &rpc.CallOption{
			Timeout:      time.Second * 10,
			Async:        true,
			AsyncRetChan: r.rpcReplyChan,
			ReqMd:        md,
			PassThrough:  &RoutingRpcPassThrough{Cli: client, ReqMsgName: wsMessage.MsgName, Uuid: wsMessage.Uuid, Service: service, Method: method},
		}
		if _, _, _err := cc.Invoke(ctx, service, method, wsMessage.Payload, nil, callOpt); _err != nil {
			mlog.Error("routing invoke error: %v", _err)
			return nil, _err
		}
		return nil, nil
	})
}

type RoutingRpcPassThrough struct {
	Cli        *WsClient
	ReqMsgName string
	Uuid       string
	Service    string
	Method     string
}

func (r *RoutingWorkerImp) ProcessRpcReply(rpcReply *rpc.AsyncCallResult) {
	passData := rpcReply.PassThrough.(*RoutingRpcPassThrough)
	if passData.ReqMsgName == "game.CLogin" {
		cli := passData.Cli
		cli.PlayerId = rpcReply.RspMd.GetInt(values.Rpc_SessionId)
		ClientMgr.AddClient(passData.Cli)
	}
	msgName := passData.Service + ".S" + passData.Method
	rspData := rpcReply.Rsp.([]byte)
	replyClientResponse(passData.Cli, passData.Uuid, msgName, rpcReply.RspMd, rspData, rpcReply.Err)
}

func replyClientResponse(cli *WsClient, uuid, msgName string, rspMd *rpc.Meta, rspData []byte, callErr error) {
	wsRsp := &ws.WsResponseMessage{Uuid: uuid}
	if callErr != nil {
		codeErr, ok := callErr.(errs.CodeError)
		if ok {
			wsRsp.ErrorCode = codeErr.Code()
			wsRsp.ErrorDesc = codeErr.Error()
		} else {
			wsRsp.ErrorCode = 1
			wsRsp.ErrorDesc = callErr.Error()
		}
	} else {
		wsRsp.MsgName = msgName
		if offsets := rspMd.IntValues(values.Rpc_NoticeOffset); len(offsets) > 0 {
			// response
			wsRsp.Payload = rspData[:offsets[0]]
			// notices
			for i := 0; i < len(offsets); i++ {
				wsNotice := &ws.PBPackage{}
				var pushData []byte
				if i == len(offsets)-1 {
					pushData = rspData[offsets[i]:]
				} else {
					pushData = rspData[offsets[i]:offsets[i+1]]
				}
				if err := proto.Unmarshal(pushData, wsNotice); err != nil {
					mlog.Error("Unmarshal ws.PBPackage error: %v", err)
					return
				}
				wsRsp.Notices = append(wsRsp.Notices, wsNotice)
			}
		} else {
			wsRsp.Payload = rspData
		}
	}

	datas, err := proto.Marshal(wsRsp)
	if err != nil {
		mlog.Error("marshal wsRsp error: %v", err)
		return
	}
	err = cli.conn.Send(datas)
	if err != nil {
		mlog.Error("send wsRsp error: %v", err)
	}
}

func getServiceNodeName(cli *WsClient, service string) string {
	switch service {
	case "game":
		return fmt.Sprintf("game.%d", 1) //fmt.Sprintf("game%d", cli.ServerId)
	default:
		return service
	}
}

func (r *RoutingWorkerImp) PushData(session any, datas []byte) {
	r.routingChan <- &RoutingTask{
		Cli:   session.(*WsClient),
		Datas: datas,
	}
}

func (r *RoutingWorkerImp) Run(wg *sync.WaitGroup, quit chan struct{}) {
	defer func() {
		wg.Done()
		mlog.Info("RoutingWorkerImp exit")
	}()
	for {
		select {
		case <-quit:
			return
		case task := <-r.routingChan:
			r.RoutingMsg(task)
		case rpcReply := <-r.rpcReplyChan:
			r.ProcessRpcReply(rpcReply)
		}
	}
}

type _LoadBalanceImp struct {
	workerSize int
	taskSize   int
	workers    []wsg.RoutingWorker
	cur        atomic.Uint32
	wg         sync.WaitGroup
	quit       chan struct{}
}

func newRouterPool(workerSize, taskSize int) *_LoadBalanceImp {
	p := &_LoadBalanceImp{
		workerSize: workerSize,
		taskSize:   taskSize,
		quit:       make(chan struct{}),
	}
	return p
}

func (p *_LoadBalanceImp) Start() {
	p.workers = make([]wsg.RoutingWorker, p.workerSize)
	for i := 0; i < p.workerSize; i++ {
		worker := &RoutingWorkerImp{
			routingChan:  make(chan *RoutingTask, p.taskSize),
			rpcReplyChan: make(chan *rpc.AsyncCallResult, p.taskSize),
		}
		p.workers[i] = worker
		p.wg.Add(1)
		go worker.Run(&p.wg, p.quit)
	}
}

func (p *_LoadBalanceImp) Stop() {
	close(p.quit)
	p.wg.Wait()
}

func (p *_LoadBalanceImp) GetOne(cli *WsClient) wsg.RoutingWorker {
	// 轮询
	idx := int(p.cur.Add(1)) % p.workerSize
	return p.workers[idx]
}

func (p *_LoadBalanceImp) OnHandshake(conn *wsg.Conn, req *http.Request) error {
	// fmt.Printf("URL: %v\n", req.URL.String())
	// fmt.Printf("Header: %v\n", req.Header)
	cli := &WsClient{
		conn: conn,
		// Account:  req.Header.Get("x-account"),
		// PlayerId: wsg.HttpHeaderGetInt64(req.Header, "x-player-id"),
		// ServerId: wsg.HttpHeaderGetInt64(req.Header, "x-server-id"),
	}
	conn.BindSession(cli)
	router := p.GetOne(cli)
	conn.BindRoutingWorker(router)

	mlog.Info("player %d Handshake ok", cli.PlayerId)
	return nil
}
