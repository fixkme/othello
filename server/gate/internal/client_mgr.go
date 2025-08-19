package internal

import (
	"context"
	"sync"

	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/gokit/wsg"
	"github.com/fixkme/othello/server/common/const/values"
	"github.com/fixkme/othello/server/common/shared"
	"github.com/fixkme/othello/server/pb/game"
	"google.golang.org/protobuf/proto"
)

type WsClient struct {
	conn *wsg.Conn

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
	pid := cli.PlayerId
	mlog.Info("player ws closed, acc:%s, pid:%d, addr:%s, reason:%v", cli.Account, pid, conn.RemoteAddr().String(), err)
	if pid > 0 {
		ClientMgr.RemoveClient(pid)
		// 通知game玩家下线
		gameServiceNode := getServiceNodeName(cli, values.Service_Game)
		_, callErr := RpcModule.GetRpcImp().Call(gameServiceNode, func(ctx context.Context, cc *rpc.ClientConn) (proto.Message, error) {
			_err := shared.AsyncCall(ctx, cc, &game.CPlayerOffline{PlayerId: pid})
			return nil, _err
		})
		if callErr != nil {
			mlog.Error("player %d call PlayerOffline failed, %v", pid, callErr)
		}
	}
}
