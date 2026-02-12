package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fixkme/gokit/errs"
	"github.com/fixkme/gokit/framework/core"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/gokit/wsg"
	"github.com/fixkme/othello/server/common/values"
	"github.com/fixkme/othello/server/pb/ws"
	"google.golang.org/protobuf/proto"
)

const (
	cloginMsgName     = "hall.CLogin"
	centerGameMsgName = "hall.CEnterGame"
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
	defer func() {
		if err := recover(); err != nil {
			mlog.Errorf("RoutingWorkerImp RoutingMsg panic: %v\n%s", err, debug.Stack())
		}
	}()
	client := task.Cli
	// 反序列化数据
	wsMessage := &ws.WsRequestMessage{}
	err := proto.Unmarshal(task.Datas, wsMessage)
	if err != nil {
		mlog.Errorf("proto.Unmarshal wsReq err:%v", err)
		client.conn.Close()
		return
	}
	// 检查是否登录
	if client.PlayerId == 0 && wsMessage.MsgName != cloginMsgName {
		mlog.Errorf("first msg must is %s", cloginMsgName)
		client.conn.Close()
		return
	}

	defer func() {
		if r := recover(); r != nil {
			mlog.Errorf("routing player %s,%d panic: %v", client.Account, client.PlayerId, r)
			err = errors.New("routing panic")
		}
		if err != nil {
			mlog.Errorf("routing player %s,%d error: %v", client.Account, client.PlayerId, err)
			respMsgName := strings.Replace(wsMessage.MsgName, ".C", ".S", 1)
			replyClientResponse(client, wsMessage.Uuid, respMsgName, nil, nil, err)
			if wsMessage.MsgName == cloginMsgName {
				client.conn.Close()
			}
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
	_, err = core.Rpc.Call(serviceNode, func(ctx context.Context, cc *rpc.ClientConn) (proto.Message, error) {
		md := &rpc.Meta{}
		md.AddStr(values.Rpc_GateId, GateNodeId)
		if client.PlayerId != 0 {
			md.AddInt(values.Rpc_PlayerId, client.PlayerId)
		}
		callOpt := &rpc.CallOption{
			Timeout:      time.Second * 10,
			Async:        true,
			AsyncRetChan: r.rpcReplyChan,
			ReqMd:        md,
			PassThrough:  &RoutingRpcPassThrough{Cli: client, ReqMsgName: wsMessage.MsgName, Uuid: wsMessage.Uuid, Service: service, Method: method},
		}
		if _, _, _err := cc.Invoke(ctx, service, method, wsMessage.Payload, nil, callOpt); _err != nil {
			mlog.Errorf("routing invoke error: %v", _err)
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
	defer func() {
		if err := recover(); err != nil {
			mlog.Errorf("RoutingWorkerImp ProcessRpcReply panic: %v\n%s", err, debug.Stack())
		}
	}()
	passData := rpcReply.PassThrough.(*RoutingRpcPassThrough)
	//respMsgName := strings.Replace(passData.ReqMsgName, ".C", ".S", 1)
	respMsgName := passData.Service + ".S" + passData.Method
	rpcErr := rpcReply.Err
	if rpcErr == nil {
		careMsg(rpcReply)
	}
	rspData, _ := rpcReply.Rsp.([]byte)
	replyClientResponse(passData.Cli, passData.Uuid, respMsgName, rpcReply.RspMd, rspData, rpcErr)
}

func replyClientResponse(cli *WsClient, uuid, msgName string, rspMd *rpc.Meta, rspData []byte, callErr error) {
	mlog.Debugf("ProcessRpcReply replyClientResponse:%s, msg:%s, callErr:%v, rspDataSize:%d", cli.Account, msgName, callErr, len(rspData))
	wsRsp := &ws.WsResponseMessage{Uuid: uuid, MsgName: msgName}
	if callErr != nil {
		codeErr, ok := callErr.(errs.CodeError)
		if ok {
			wsRsp.ErrorCode = codeErr.Code()
			wsRsp.ErrorDesc = codeErr.Error()
		} else {
			wsRsp.ErrorCode = errs.ErrCode_Unknown
			wsRsp.ErrorDesc = callErr.Error()
		}
	} else {
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
					mlog.Errorf("Unmarshal ws.PBPackage error: %v", err)
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
		mlog.Errorf("marshal wsRsp error: %v", err)
		return
	}
	err = cli.conn.Send(datas)
	if err != nil {
		mlog.Errorf("send wsRsp error: %v", err)
	}
}

func getServiceNodeName(cli *WsClient, service string) string {
	switch service {
	case "game":
		return fmt.Sprintf("game.%d", cli.GameId)
	default:
		return service
	}
}

func careMsg(rpcReply *rpc.AsyncCallResult) {
	passData := rpcReply.PassThrough.(*RoutingRpcPassThrough)
	cli := passData.Cli
	switch passData.ReqMsgName {
	case cloginMsgName:
		cli.PlayerId = rpcReply.RspMd.GetInt(values.Rpc_PlayerId)
		if cli.PlayerId > 0 {
			ClientMgr.AddClient(passData.Cli)
		} else {
			mlog.Errorf("careMsg slogin no session id, acc:%s, rspMd:%v", cli.Account, rpcReply.RspMd)
		}
	case centerGameMsgName:
		cli.GameId = rpcReply.RspMd.GetInt(values.Rpc_GameId)
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
		mlog.Infof("RoutingWorkerImp exit")
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
	params := req.URL.Query()
	mlog.Debugf("OnHandshake http.Request params: %v", params)

	cli := &WsClient{
		conn:    conn,
		Account: params.Get("x-account"),
		// PlayerId: wsg.HttpHeaderGetInt64(req.Header, "x-player-id"),
		// ServerId: wsg.HttpHeaderGetInt64(req.Header, "x-server-id"),
	}
	conn.BindSession(cli)
	router := p.GetOne(cli)
	conn.BindRoutingWorker(router)

	mlog.Infof("player [%v] Handshake ok", cli.Account)
	return nil
}
