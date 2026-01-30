package ui

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/fixkme/othello/server/prototesting/net"
	"github.com/fixkme/othello/server/prototesting/pb"
	"google.golang.org/protobuf/proto"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type ProtoWidget struct {
	Container          *fyne.Container
	MsgTypeInput       *widget.Select
	MsgTypeFilterInput *widget.Entry
	MsgGenerator       *widget.Button
	MsgSender          *widget.Button
	MsgBackward        *widget.Button
	MsgForward         *widget.Button
	ProtoEditor        *widget.Entry
	parent             fyne.Window
	histories          []*MsgCommand
	historyCursor      int

	client          *net.Client
	selectedMsgType string
}

func NewProtoWidget(p fyne.Window, cli *net.Client) *ProtoWidget {
	wproto := &ProtoWidget{
		parent:    p,
		client:    cli,
		histories: make([]*MsgCommand, 0, 16),
	}
	msgTypes := pb.GetAllRequestType()
	wproto.MsgTypeInput = widget.NewSelect(msgTypes, wproto.selectMsgType)
	wproto.MsgTypeFilterInput = widget.NewEntry()
	wproto.MsgTypeFilterInput.OnChanged = wproto.filterMsgType
	wproto.MsgTypeFilterInput.SetPlaceHolder("消息过滤字符串, 大小写不敏感")

	wproto.MsgGenerator = widget.NewButton("生成", wproto.generateMsg)
	wproto.MsgSender = widget.NewButton("发送", wproto.sendMsg)
	wproto.MsgBackward = widget.NewButton("后退", wproto.msgBackward)
	wproto.MsgForward = widget.NewButton("前进", wproto.msgForward)
	hbox := container.NewHBox(wproto.MsgGenerator, wproto.MsgSender, wproto.MsgBackward, wproto.MsgForward)
	topBox := container.NewBorder(nil, nil, widget.NewLabel("消息类型:"), hbox, wproto.MsgTypeInput)
	msgTypeLabelEntry := container.NewBorder(nil, nil, widget.NewLabel("搜索消息类型:"), nil, wproto.MsgTypeFilterInput)
	secBox := container.NewVBox(msgTypeLabelEntry)

	wproto.ProtoEditor = widget.NewMultiLineEntry()
	topContainer := container.NewVBox(topBox, secBox)
	wproto.Container = container.NewBorder(topContainer, nil, nil, nil, wproto.ProtoEditor)

	wproto.MsgBackward.Disable()
	wproto.MsgForward.Disable()

	return wproto
}

func (w *ProtoWidget) selectMsgType(msgType string) {
	fmt.Printf("select msgType %s\n", msgType)
	w.selectedMsgType = msgType
}

func (w *ProtoWidget) fillFields(msg proto.Message) {
	t := reflect.TypeOf(msg).Elem()
	v := reflect.ValueOf(msg).Elem()
	for i := 0; i < t.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := t.Field(i).Type
		name := t.Field(i).Name
		if unicode.IsLower(rune(name[0])) {
			continue
		}
		if _, ok := fieldValue.Interface().(proto.Message); ok {
			ft := fieldType.Elem()
			fv := reflect.New(ft)
			fieldValue.Set(fv)
			w.fillFields(fv.Interface().(proto.Message))
		} else {
			if fieldType.Kind() == reflect.Slice {
				elemType := fieldType.Elem()
				if elemType.Kind() != reflect.Ptr {
					continue
				}
				newElement := reflect.New(elemType.Elem()).Interface()
				ev := reflect.ValueOf(newElement)
				if _, ok := ev.Interface().(proto.Message); ok {
					fieldValue.Set(reflect.Append(fieldValue, ev))
					w.fillFields(ev.Interface().(proto.Message))
				}
			}
		}
	}
}

func (w *ProtoWidget) filterMsgType(filterStr string) {
	msgTypes := pb.GetAllRequestType()
	subStr := strings.ToLower(strings.TrimSpace(filterStr))
	if len(subStr) == 0 {
		if len(msgTypes) == len(w.MsgTypeInput.Options) {
			return
		}
		options := make([]string, len(msgTypes))
		copy(options, msgTypes)
		w.MsgTypeInput.SetOptions(options)
		w.MsgTypeInput.Refresh()
	} else {
		options := make([]string, 0, len(msgTypes))
		for _, s := range msgTypes {
			if strings.Contains(strings.ToLower(s), subStr) {
				options = append(options, s)
			}
		}
		w.MsgTypeInput.SetOptions(options)
		w.MsgTypeInput.Refresh()
	}
}

func (w *ProtoWidget) generateMsg() {
	if len(w.selectedMsgType) == 0 {
		return
	}
	reqMsg := pb.NewRequestMessage(w.selectedMsgType)
	if reqMsg == nil {
		dialog.ShowInformation("", "此消息没有注册", w.parent)
		return
	}
	w.fillFields(reqMsg)
	text, err := pb.PbToJsonText(reqMsg)
	if err != nil {
		errstr := fmt.Sprintf("generateMsg PbToJsonText err:%v", err)
		dialog.ShowInformation("", errstr, w.parent)
		return
	}
	w.ProtoEditor.SetText(text)
}

func (w *ProtoWidget) sendMsg() {
	if len(w.selectedMsgType) == 0 {
		return
	}
	reqMsg := pb.NewRequestMessage(w.selectedMsgType)
	if reqMsg == nil {
		dialog.ShowInformation("", "此消息没有注册", w.parent)
		return
	}
	if err := pb.JsonTextToPb(w.ProtoEditor.Text, reqMsg); err != nil {
		errstr := fmt.Sprintf("sendMsg JsonTextToPb err:%v", err)
		dialog.ShowInformation("", errstr, w.parent)
		return
	}
	if w.client.SendMsg(reqMsg) {
		w.updateCommandHistory()
	}
}

func (w *ProtoWidget) updateCommandHistory() {
	if len(w.histories) > 0 {
		last := w.histories[len(w.histories)-1]
		if last.MsgType == w.selectedMsgType && last.MsgContent == w.ProtoEditor.Text {
			return
		}
	}
	w.histories = append(w.histories, &MsgCommand{MsgType: w.selectedMsgType, MsgContent: w.ProtoEditor.Text})
	if len(w.histories) > 100 {
		w.histories = w.histories[len(w.histories)-100:]
	}
	w.historyCursor = len(w.histories) - 1
	if len(w.histories) > 1 {
		w.MsgBackward.Enable()
	}
	w.MsgForward.Disable()
}

func (w *ProtoWidget) msgBackward() {
	if w.historyCursor <= 0 {
		return
	}
	cmd := w.histories[w.historyCursor-1]
	w.MsgTypeInput.SetSelected(cmd.MsgType)
	w.ProtoEditor.SetText(cmd.MsgContent)
	w.MsgForward.Enable()
	w.historyCursor--
	if w.historyCursor <= 0 {
		w.MsgBackward.Disable()
	}
}

func (w *ProtoWidget) msgForward() {
	if w.historyCursor >= len(w.histories)-1 {
		return
	}
	cmd := w.histories[w.historyCursor+1]
	w.MsgTypeInput.SetSelected(cmd.MsgType)
	w.ProtoEditor.SetText(cmd.MsgContent)
	w.MsgBackward.Enable()
	w.historyCursor++
	if w.historyCursor >= len(w.histories)-1 {
		w.MsgForward.Disable()
	}
}

type MsgCommand struct {
	MsgType    string
	MsgContent string
}
