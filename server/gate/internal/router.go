package internal

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/fixkme/gokit/wsg"
)

type RoutingTask struct {
	Cli   *WsClient
	Datas []byte
}

type RoutingWorkerImp struct {
	tasks chan *RoutingTask
}

func (r *RoutingWorkerImp) DoTask(task *RoutingTask) {
	// 反序列化数据
	// 这里只是用echo作为测试
	//fmt.Printf("routing data %d:%v\n", len(task.Datas), string(task.Datas))
	conn := task.Cli.conn
	conn.Send(task.Datas)
	// 路由数据到game或其他

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
		log.Println("RoutingWorkerImp exit")
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
	fmt.Println("Handshake ok ", cli.PlayerId)
	return nil
}

type WsClient struct {
	conn *wsg.Conn

	Account  string
	PlayerId int64
	ServerId int64
}
