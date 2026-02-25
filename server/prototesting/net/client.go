package net

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/fixkme/othello/server/pb/ws"
	"github.com/fixkme/othello/server/prototesting/conf"
	"github.com/fixkme/othello/server/prototesting/pb"
	"github.com/gorilla/websocket"
)

const (
	DisConnect = 0
	Connecting = 1
	Connected  = 2
)

type Client struct {
	account   string
	playerId  int64
	serverId  int64
	appSecret string

	conn         *websocket.Conn
	status       int
	closed       atomic.Bool
	disconnectCb func(error)
	sendChan     chan proto.Message
	msgChan      chan *MessageItem
	ctx          context.Context
	cancelFunc   context.CancelFunc
}

func NewClient() *Client {
	cli := &Client{
		account:   "acc_test",
		playerId:  101,
		serverId:  1,
		appSecret: "foobar",
		sendChan:  make(chan proto.Message, 16),
		msgChan:   make(chan *MessageItem, 1024),
	}
	cli.closed.Store(true)
	return cli
}

func (cli *Client) Status() int {
	return cli.status
}

func (cli *Client) GetMsgReader() <-chan *MessageItem {
	return cli.msgChan
}

func (cli *Client) SetServerId(serverId int64) {
	cli.serverId = serverId
}
func (cli *Client) SetPlayerId(playerId int64) {
	cli.playerId = playerId
}
func (cli *Client) GetPlayerId() int64 {
	return cli.playerId
}
func (cli *Client) SetAccount(acc string) {
	cli.account = acc
}
func (cli *Client) GetAccount() string {
	return cli.account
}

func (cli *Client) SetDisconnectCb(f func(error)) {
	cli.disconnectCb = f
}

func (cli *Client) ConnectHost(hostName string) error {
	header := http.Header{}
	header.Set("x-debug-player-id", strconv.Itoa(int(cli.playerId)))
	header.Set("x-encrypted", "true")
	header.Set("x-gzip", "true")
	header.Set("x-typ", "protobuf")
	header.Set("x-server-id", strconv.Itoa(int(cli.serverId)))
	header.Set("x-character-id", strconv.Itoa(int(cli.playerId)))

	svr := conf.ServerList[hostName]
	u := url.URL{Scheme: svr.Scheme, Host: svr.Host, Path: svr.Path, RawQuery: fmt.Sprintf("x-account=%s", cli.account)}
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 10 * time.Second,
	}
	fmt.Printf("dial url %v, header %v\n", u.String(), header)
	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		return err
	}
	cli.conn = conn
	cli.status = Connected
	cli.closed.Store(false)
	cli.ctx, cli.cancelFunc = context.WithCancel(context.Background())
	go cli.sendLoop()
	go cli.recvLoop()
	return nil
}

func (cli *Client) SendMsg(msg proto.Message) bool {
	if cli.closed.Load() || cli.status != Connected {
		return false
	}
	cli.sendChan <- msg
	return true
}

func (cli *Client) Stop() {
	cli.cancelFunc()
}

func (cli *Client) closeConn() {
	if !cli.closed.Load() {
		cli.conn.Close()
		cli.closed.Store(true)
	}
}

func (cli *Client) sendLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("send panic: %v\n", r)
		}
		cli.closeConn()
	}()
	heartBeatTime := 20 * time.Second
	timer := time.NewTimer(heartBeatTime)
	for {
		select {
		case <-cli.ctx.Done():
			return
		case <-timer.C:
			if err := cli.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("ping err:%v\n", err)
				cli.cancelFunc()
				if cli.disconnectCb != nil {
					cli.disconnectCb(err)
				}
			}
			timer.Reset(heartBeatTime)
		case msg, ok := <-cli.sendChan:
			if ok {
				if err := cli.send(msg); err != nil {
					log.Printf("send err:%v\n", err)
					cli.cancelFunc()
					if cli.disconnectCb != nil {
						cli.disconnectCb(err)
					}
					return
				}
			} else {
				cli.cancelFunc()
				return
			}
		}
	}
}

func (cli *Client) recvLoop() {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		log.Printf("recv panic: %v\n", r)
	// 	}
	// 	cli.closeConn()
	// }()
	for {
		select {
		case <-cli.ctx.Done():
			return
		default:
			if err := cli.recv(); err != nil {
				log.Printf("recv err:%v\n", err)
				cli.cancelFunc()
				if cli.disconnectCb != nil {
					cli.disconnectCb(err)
				}
				return
			}
		}
	}
}

func (cli *Client) send(req proto.Message) error {
	var err error
	wsMsg := &ws.WsRequestMessage{}
	msgItem := makeReqMessageItem(req)
	wsMsg.MsgName = string(req.ProtoReflect().Descriptor().FullName())
	wsMsg.Uuid = msgItem.Id
	bytes, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	wsMsg.Payload = bytes
	data, err := proto.Marshal(wsMsg)
	if err != nil {
		return err
	}

	if err = cli.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return err
	}
	cli.msgChan <- msgItem
	return nil
}

func (cli *Client) recv() error {
	// wsh, err := wsg.ReadWsHeader(cli.conn.NetConn())
	// if err != nil {
	// 	return err
	// }
	// messageType := wsh.OpCode
	// bytes := make([]byte, wsh.Length)
	// n, err := io.ReadFull(cli.conn.NetConn(), bytes)
	// if err != nil {
	// 	return err
	// }
	messageType, bytes, err := cli.conn.ReadMessage()
	if err != nil {
		return err
	}
	if messageType == websocket.BinaryMessage {
		wsMsg := &ws.WsResponseMessage{}
		if err = proto.Unmarshal(bytes, wsMsg); err != nil {
			return err
		}
		if len(wsMsg.Uuid) > 0 {
			log.Printf("recv response: %s, %s\n", wsMsg.Uuid, wsMsg.MsgName)
			if wsMsg.ErrorCode == 0 {
				msgType, msgData := wsMsg.MsgName, wsMsg.Payload
				protoMsg, err := cli.decodeRecvMessage(msgType, msgData, cli.appSecret, true, true, false)
				if err != nil {
					return err
				}
				cli.msgChan <- makeRspMessageItem(protoMsg, wsMsg.Uuid)
			} else {
				var name string = wsMsg.MsgName
				cli.msgChan <- makeErrorMessageItem(name, wsMsg.Uuid, wsMsg.ErrorCode, wsMsg.ErrorDesc)
			}
		}
		if len(wsMsg.Notices) > 0 {
			for idx, ntfMsg := range wsMsg.Notices {
				log.Printf("recv push: %s, %s\n", wsMsg.Uuid, ntfMsg.MessageType)
				msgType, msgData := ntfMsg.MessageType, ntfMsg.MessagePayload
				protoMsg, err := cli.decodeRecvMessage(msgType, msgData, cli.appSecret, true, true, true)
				if err != nil {
					return err
				}
				cli.msgChan <- makeNoticeMessageItem(protoMsg, wsMsg.Uuid, idx)
			}
		}
	} else {
		log.Println("ws op message type error")
	}
	return nil
}

func (cli *Client) decodeRecvMessage(messageType string, raw []byte, secret string, encrypt, compress, notice bool) (proto.Message, error) {
	// payload, err := encoding.DecodeRaw(raw, secret, encrypt, compress)
	// if err != nil {
	// 	return nil, err
	// }
	payload := raw
	var protoMsg protoreflect.ProtoMessage
	if notice {
		protoMsg = pb.NewNoticeMessage(messageType)
	} else {
		protoMsg = pb.NewResponseMessage(messageType)
	}
	if protoMsg == nil {
		return nil, fmt.Errorf("messageType (%s) not register", messageType)
	}
	if err := proto.Unmarshal(payload, protoMsg); err != nil {
		return nil, err
	}
	return protoMsg, nil
}

func (cli *Client) getMsgUri(msg proto.Message) (string, bool) {
	return "", true
}
