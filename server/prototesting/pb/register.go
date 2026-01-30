package pb

import (
	"fmt"
	"sort"

	"github.com/fixkme/gokit/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var reqMsgMap map[string]proto.Message
var rspMsgMap map[string]proto.Message
var noticeMsgMap map[string]proto.Message

func RegisterRequestMessage(msg proto.Message) {
	name := util.GetMessageFullName(msg)
	if reqMsgMap == nil {
		reqMsgMap = make(map[string]proto.Message)
	}
	reqMsgMap[name] = msg
}

func RegisterResponseMessage(msg proto.Message) {
	name := util.GetMessageFullName(msg)
	if rspMsgMap == nil {
		rspMsgMap = make(map[string]proto.Message)
	}
	rspMsgMap[name] = msg
}

func RegisterNoticeMessage(msg proto.Message) {
	name := util.GetMessageFullName(msg)
	if noticeMsgMap == nil {
		noticeMsgMap = make(map[string]proto.Message)
	}
	noticeMsgMap[name] = msg
}

func NewRequestMessage(name string) proto.Message {
	msg, ok := reqMsgMap[name]
	if ok {
		return msg.ProtoReflect().New().Interface()
	} else {
		return nil
	}
}

func NewResponseMessage(name string) proto.Message {
	msg, ok := rspMsgMap[name]
	if ok {
		return msg.ProtoReflect().New().Interface()
	} else {
		return nil
	}
}

func NewNoticeMessage(name string) proto.Message {
	msg, ok := noticeMsgMap[name]
	if ok {
		return msg.ProtoReflect().New().Interface()
	} else {
		return nil
	}
}

func RegisterMessage() {
	for _, msgFullName := range RequestMsgNames {
		mt, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(msgFullName))
		if err != nil {
			panic(fmt.Sprintf("load message %s failed: %v", msgFullName, err))
		}
		RegisterRequestMessage(mt.New().Interface())
	}
	for _, msgFullName := range ResponseMsgNames {
		mt, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(msgFullName))
		if err != nil {
			panic(fmt.Sprintf("load message %s failed: %v", msgFullName, err))
		}
		RegisterResponseMessage(mt.New().Interface())
	}
	for _, msgFullName := range NoticeMsgNames {
		mt, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(msgFullName))
		if err != nil {
			panic(fmt.Sprintf("load message %s failed: %v", msgFullName, err))
		}
		RegisterNoticeMessage(mt.New().Interface())
	}
}

var msgNames []string

func GetAllRequestType() []string {
	if len(msgNames) > 0 {
		return msgNames
	}
	types := make([]string, len(RequestMsgNames))
	copy(types, RequestMsgNames)
	sort.Strings(types)
	msgNames = types
	return types
}
