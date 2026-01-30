package net

import (
	"encoding/json"
	"fmt"

	"github.com/fixkme/gokit/util"
	"github.com/rs/xid"
	"google.golang.org/protobuf/proto"
)

const (
	MsgType_Request  = 1
	MsgType_Response = 2
	MsgType_Notice   = 3
	MsgType_Error    = 4
)

type MessageItem struct {
	Id      string
	Name    string
	Msg     proto.Message
	MsgType int
}

func (msg MessageItem) FormatString() string {
	data, _ := json.Marshal(msg.Msg)
	str := fmt.Sprintf("%s %s %s", msg.Id, msg.Name, string(data))
	return str
}

func makeReqMessageItem(msg proto.Message) *MessageItem {
	item := &MessageItem{
		Id:      xid.New().String(),
		Name:    util.GetMessageName(msg),
		Msg:     msg,
		MsgType: MsgType_Request,
	}
	return item
}

func makeRspMessageItem(msg proto.Message, id string) *MessageItem {
	item := &MessageItem{
		Id:      id,
		Name:    util.GetMessageName(msg),
		Msg:     msg,
		MsgType: MsgType_Response,
	}
	return item
}

func makeErrorMessageItem(name, id string, code int32, msg string) *MessageItem {
	codeText := fmt.Sprintf("UnknownCode(%d)", code)
	item := &MessageItem{
		Id:      id,
		Name:    fmt.Sprintf("type: %s, code: %s, msg: %s", name, codeText, msg),
		Msg:     nil,
		MsgType: MsgType_Error,
	}
	return item
}

func makeNoticeMessageItem(msg proto.Message, id string, idx int) *MessageItem {
	if len(id) == 0 {
		id = xid.New().String()
	}
	item := &MessageItem{
		Id:      fmt.Sprintf("%s-Notice-%d", id, idx+1),
		Name:    util.GetMessageName(msg),
		Msg:     msg,
		MsgType: MsgType_Notice,
	}
	return item
}
