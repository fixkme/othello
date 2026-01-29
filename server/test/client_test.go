package test

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/fixkme/gokit/wsg"
	"github.com/fixkme/othello/server/pb/hall"
	"github.com/fixkme/othello/server/pb/ws"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

func TestLogin(t *testing.T) {
	req := &hall.CLogin{Account: "acc_test"}
	rsp := &hall.SLogin{}
	for i := range 1 {
		go client(i+1, req, rsp)
	}

	select {
	case <-time.After(time.Second * 300):
	}

}

func client(pid int, msg proto.Message, out proto.Message) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	host := "127.0.0.1:7070"               // 服务端地址
	localIP := "127.0.0.1"                 // 本地IP地址
	localPort := strconv.Itoa(54680 + pid) // 本地端口

	// 解析本地地址
	localAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(localIP, localPort))
	if err != nil {
		panic(err)
	}
	// 创建自定义Dialer
	dialer := &net.Dialer{
		LocalAddr: localAddr,
		Timeout:   1 * time.Second,
	}
	conn, err := dialer.Dial("tcp", host)
	if err != nil {
		panic(err)
	}
	key, _ := generateWebSocketKey()
	contentFormat := "GET /ws?x-account=acc_test_%d HTTP/1.1\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: %s\r\n" +
		"Host: %s\r\nx-encrypted: false\r\nx-debug-player-id: %d\r\nx-server-id: 1\r\nx-character-id: %d\r\n"
	content := fmt.Sprintf(contentFormat, pid, key, host, pid, pid)
	if _, err = conn.Write([]byte(content)); err != nil {
		panic(err)
	}
	//time.Sleep(time.Second * time.Duration(10))
	if _, err = conn.Write([]byte("\r\n")); err != nil {
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

	// 慢慢发送数据
	// for _, c := range hbf {
	// 	if _, err = conn.Write([]byte{c}); err != nil {
	// 		fmt.Println(err)
	// 		return
	// 	}
	// 	time.Sleep(time.Millisecond * 10)
	// }
	// for _, c := range pd {
	// 	if _, err = conn.Write([]byte{c}); err != nil {
	// 		fmt.Println(err)
	// 		return
	// 	}
	// 	time.Sleep(time.Millisecond * 10)
	// }

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
	os.Exit(0)
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
