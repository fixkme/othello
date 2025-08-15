package test

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/fixkme/gokit/wsg"
	"github.com/fixkme/othello/server/pb/game"
	"github.com/fixkme/othello/server/pb/ws"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

func TestLogin(t *testing.T) {
	req := &game.CLogin{Account: "acc_test"}
	rsp := &game.SLogin{}
	go client(req, rsp)
	<-time.After(time.Second * 10)
}

func client(msg proto.Message, out proto.Message) {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:7070", time.Second)
	if err != nil {
		panic(err)
	}
	key, _ := generateWebSocketKey()
	content := "GET /chat HTTP/1.1\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: " + key + "\r\n" +
		"x-player-id: " + strconv.Itoa(10) + "\r\n\r\n"
	if _, err = conn.Write([]byte(content)); err != nil {
		panic(err)
	}
	//fmt.Printf("发送握手请求报文完毕\n")
	fmt.Println(string(readHttpReply(conn)))

	msgName := string(msg.ProtoReflect().Descriptor().FullName())
	payload, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	wsReq := &ws.WsRequestMessage{Uuid: "1", MsgName: msgName, Payload: payload}
	wsData, err := proto.Marshal(wsReq)
	if err != nil {
		panic(err)
	}
	pd := wsData
	wsh := &wsg.WsHead{Fin: true, OpCode: wsg.OpBinary, Masked: true, Length: int64(len(pd))}
	geneMask(&wsh.Mask, pd)
	hbf, _ := wsg.MakeWsHeadBuff(wsh)
	//fmt.Println(wsh.Mask, ",", string(pd))
	if _, err = conn.Write(hbf); err != nil {
		fmt.Println(err)
		return
	}
	if _, err = conn.Write(pd); err != nil {
		fmt.Println(err)
		return
	}

	wsRsp := &ws.WsResponseMessage{}
	wsData = readWsReply(conn)
	err = proto.Unmarshal(wsData, wsRsp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("wsRsp:", wsRsp.String())
	if wsRsp.ErrorCode == 0 {
		err = proto.Unmarshal(wsRsp.Payload, out)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(msgName, ":", prototext.Format(out))
	}
}

// 客户端生成 WebSocket key
func generateWebSocketKey() (string, error) {
	// 生成 16 字节的随机数
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	// 进行 Base64 编码
	return base64.StdEncoding.EncodeToString(key), nil
}

// 客户端对payload进行掩码
func geneMask(mask *[4]byte, payload []byte) {
	k := (*mask)[:]
	_, err := rand.Read(k)
	if err != nil {
		panic(err)
	}
	wsg.MaskWsPayload(*mask, payload)
}

func readHttpReply(c net.Conn) []byte {
	rpy := make([]byte, 1024)
	n, err := c.Read(rpy)
	if err != nil {
		panic(err)
	}
	return rpy[:n]
}

func readWsReply(c net.Conn) []byte {
	wsh, err := wsg.ReadWsHeader(c)
	if err != nil {
		panic(err)
	}
	payload := make([]byte, wsh.Length)
	n, err := io.ReadFull(c, payload)
	if err != nil {
		panic(err)
	}
	_ = n
	//fmt.Println("收到字节长度", n)
	return payload
}
