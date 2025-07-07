package internal

import (
	"sync"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/wsg"
)

type WsClient struct {
	conn      *wsg.Conn
	msgWorker *RoutingWorkerImp

	Account  string
	PlayerId int64
	ServerId int64
}

type ClientManager struct {
	clients map[int64]*WsClient
	mtx     sync.Mutex
}

var ClientMgr *ClientManager

func init() {
	ClientMgr = &ClientManager{
		clients: make(map[int64]*WsClient),
	}
}

func (m *ClientManager) AddClient(client *WsClient) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.clients[client.PlayerId] = client
}

func (m *ClientManager) RemoveClient(playerId int64) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	delete(m.clients, playerId)
}

func (m *ClientManager) GetClient(playerId int64) *WsClient {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return m.clients[playerId]
}

func OnClientClose(conn *wsg.Conn, err error) {
	cli := conn.GetSession().(*WsClient)
	mlog.Info("player %d ws closed, %v", cli.PlayerId, conn.RemoteAddr().String())
	ClientMgr.RemoveClient(cli.PlayerId)
}
