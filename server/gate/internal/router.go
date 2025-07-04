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
	"github.com/fixkme/othello/server/pb/ws"
	"google.golang.org/protobuf/proto"
)

type RoutingTask struct {
	Cli   *WsClient
	Datas []byte
}

type RoutingWorkerImp struct {
	tasks chan *RoutingTask
}

func (r *RoutingWorkerImp) DoTask(task *RoutingTask) {
	client := task.Cli
	// 反序列化数据
	wsMessage := &ws.WsRequestMessage{}
	err := proto.Unmarshal(task.Datas, wsMessage)
	if err != nil {
		mlog.Error("proto.Unmarshal wsReq err:%v", err)
		client.conn.Close()
		return
	}

	defer func() {
		if r := recover(); r != nil {
			mlog.Error("routing player %d panic %v", client.PlayerId, r)
			err = errors.New("routing panic")
		}
		if err != nil {
			replyClientResponse(client, wsMessage.Uuid, nil, nil, err)
		}
	}()

	v2 := strings.SplitN(wsMessage.MsgName, ".", 2)
	if len(v2) != 2 {
		err = fmt.Errorf("msgName invalid")
		return
	}
	service, method := v2[0], v2[1][1:]
	serviceNode := getServiceName(client, service)
	// 路由数据到game或其他
	RpcModule.GetRpcImp().Call(serviceNode, func(ctx context.Context, cc *rpc.ClientConn) (proto.Message, error) {
		callOpt := &rpc.CallOption{Timeout: time.Second * 5, Async: false, AsyncRetChan: nil}
		service = "Game"
		rspMd, rspData, callErr := cc.Invoke(context.Background(), service, method, wsMessage.Payload, nil, callOpt)
		replyClientResponse(client, wsMessage.Uuid, rspMd, rspData, callErr)
		return nil, nil
	})
}

func replyClientResponse(cli *WsClient, uuid string, rspMd *rpc.Meta, rspData []byte, callErr error) {
	wsRsp := &ws.WsResponseMessage{
		Uuid:    uuid,
		MsgName: "",
		Payload: rspData,
	}
	if callErr != nil {
		codeErr, ok := callErr.(errs.CodeError)
		if ok {
			wsRsp.ErrorCode = codeErr.Code()
			wsRsp.ErrorDesc = codeErr.Error()
		} else {
			wsRsp.ErrorCode = 1
			wsRsp.ErrorDesc = callErr.Error()
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

func getServiceName(cli *WsClient, service string) string {
	switch service {
	case "game":
		return fmt.Sprintf("game%d", 1)
	default:
		return service
	}
}

func (r *RoutingWorkerImp) PushData(session any, datas []byte) {
	r.tasks <- &RoutingTask{
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
		case task := <-r.tasks:
			r.DoTask(task)
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
			tasks: make(chan *RoutingTask, p.taskSize),
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
		conn:     conn,
		Account:  req.Header.Get("x-account"),
		PlayerId: wsg.HttpHeaderGetInt64(req.Header, "x-player-id"),
		ServerId: wsg.HttpHeaderGetInt64(req.Header, "x-server-id"),
	}
	conn.BindSession(cli)
	router := p.GetOne(cli)
	conn.BindRoutingWorker(router)
	ClientMgr.AddClient(cli)
	mlog.Info("player %d Handshake ok", cli.PlayerId)
	return nil
}
