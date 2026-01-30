package ui

import (
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/fixkme/othello/server/prototesting/net"
	"github.com/fixkme/othello/server/prototesting/pb"
)

type MessageWidget struct {
	Container        *fyne.Container
	ClearMsgBtn      *widget.Button
	MsgList          *widget.List
	parent           fyne.Window
	app              fyne.App
	currentDialogIdx int
	currentDialog    *MessageDialog

	dataList binding.ExternalStringList
	msgChan  <-chan *net.MessageItem
	msgMap   map[string][]*net.MessageItem
}

func NewMessageWidget(app fyne.App, p fyne.Window, msgChan <-chan *net.MessageItem) *MessageWidget {
	wmsg := &MessageWidget{
		app:              app,
		parent:           p,
		currentDialogIdx: -1,
		msgChan:          msgChan,
		msgMap:           make(map[string][]*net.MessageItem),
	}
	wmsg.dataList = binding.BindStringList(&[]string{})

	wmsg.ClearMsgBtn = widget.NewButton("清空消息", wmsg.ClearMsg)

	wmsg.MsgList = widget.NewListWithData(wmsg.dataList, wmsg.listCreateItem, wmsg.listUpdateItem)
	wmsg.MsgList.OnSelected = wmsg.onSelected
	wmsg.MsgList.OnUnselected = func(idx widget.ListItemID) {
		log.Printf("Unselected list item %d\n", idx)
	}

	wmsg.Container = container.NewBorder(wmsg.ClearMsgBtn, nil, nil, nil, wmsg.MsgList)
	go wmsg.acceptMsg()
	return wmsg
}

func (w *MessageWidget) onSelected(idx widget.ListItemID) {
	log.Printf("click list item %d\n", idx)
	//已经有窗口
	if w.currentDialogIdx >= 0 {
		if w.currentDialogIdx != idx {
			//w.currentDialog.Show()
			//w.MsgList.Select(w.currentDialogIdx)

			w.showMessageInfo(idx)
		}
		return
	}
	w.showMessageInfo(idx)
}

func (w *MessageWidget) processSpecialMsg(msgItems []*net.MessageItem) bool {

	return false
}

func (w *MessageWidget) showMessageInfo(idx int) {
	ss, err := w.dataList.GetValue(idx)
	if err != nil {
		dialog.ShowInformation("", err.Error(), w.parent)
		return
	}
	id := strings.Split(ss, " ")[0]
	msgItems := w.msgMap[id]
	if w.processSpecialMsg(msgItems) {
		return
	}
	var left, right string = "{}", "{}"
	for _, it := range msgItems {
		if it.MsgType == net.MsgType_Request {
			left, err = pb.PbToJsonText(it.Msg)
			if err == nil && it.Msg != nil {
				left = string(it.Msg.ProtoReflect().Descriptor().FullName()) + "\n" + left
			}
		} else if it.MsgType == net.MsgType_Response || it.MsgType == net.MsgType_Notice {
			right, err = pb.PbToJsonText(it.Msg)
			if err == nil && it.Msg != nil {
				right = string(it.Msg.ProtoReflect().Descriptor().FullName()) + "\n" + right
			}
		} else if it.MsgType == net.MsgType_Error {
			index := strings.Index(it.Name, " ")
			if index != -1 {
				right = it.Name[index+1:]
			}
		}
		if err != nil {
			dialog.ShowInformation("", err.Error(), w.parent)
			return
		}
	}
	if w.currentDialogIdx >= 0 {
		w.currentDialog.LeftText.SetText(left)
		w.currentDialog.RightText.SetText(right)
		w.currentDialog.Window.Show()
	} else {
		w.currentDialog = ShowMessageDislog(w.app, w.parent, left, right, func() {
			w.MsgList.Unselect(w.currentDialogIdx)
			w.currentDialogIdx = -1
			w.currentDialog = nil
		})
	}
	w.currentDialogIdx = idx
}

func (w *MessageWidget) acceptMsg() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MessageWidget acceptMsg panic: %v\n", r)
		}
	}()
	lastReqIndex := 0
	for {
		select {
		case msgItem, ok := <-w.msgChan:
			if !ok {
				return
			}
			str := msgItem.FormatString()
			if len(str) > 100 {
				str = str[:100]
			}
			if msgItem.MsgType == net.MsgType_Request {
				lastReqIndex = w.MsgList.Length()
				w.AppendMessage(str)
			} else if msgItem.MsgType == net.MsgType_Response || msgItem.MsgType == net.MsgType_Error {
				w.dataList.SetValue(lastReqIndex, str)
			} else {
				w.AppendMessage(str)
			}
			w.msgMap[msgItem.Id] = append(w.msgMap[msgItem.Id], msgItem)
		}
	}
}

func (w *MessageWidget) AppendMessage(content string) {
	w.dataList.Append(content)
	//w.MsgList.ScrollToBottom()
}

func (w *MessageWidget) ListLength() int {
	return w.dataList.Length()
}

func (w *MessageWidget) listCreateItem() fyne.CanvasObject {
	label := widget.NewLabel("temp")

	return label
}

func (w *MessageWidget) listUpdateItem(i binding.DataItem, o fyne.CanvasObject) {
	o.(*widget.Label).Bind(i.(binding.String))
}

func (w *MessageWidget) ClearMsg() {
	w.dataList.Set([]string{})
	w.msgMap = make(map[string][]*net.MessageItem)
}
