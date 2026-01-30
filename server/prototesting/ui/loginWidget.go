package ui

import (
	"fmt"
	"slices"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/fixkme/othello/server/pb/hall"
	"github.com/fixkme/othello/server/prototesting/conf"
	"github.com/fixkme/othello/server/prototesting/net"
)

const (
	status_DisConnect = 0
	status_Connecting = 1
	status_Connected  = 2
)

type LoginWidget struct {
	parent       fyne.Window
	Container    *fyne.Container
	HostSelect   *widget.Select
	ServerSelect *NumericalEntry
	AccountInput *widget.Entry
	LoginBtn     *widget.Button
	PlayerLabel  *widget.Label

	client      *net.Client
	status      int
	currentHost string
}

var afterLogin func(*LoginWidget)

func NewLoginWidget(p fyne.Window, cli *net.Client) *LoginWidget {
	wlogin := &LoginWidget{
		parent: p,
		client: cli,
		status: status_DisConnect,
	}

	hosts := conf.GetAllHostName()
	wlogin.HostSelect = widget.NewSelect(hosts, wlogin.hostSelected)
	wlogin.ServerSelect = NewNumericalEntry()

	wlogin.AccountInput = widget.NewEntry()
	wlogin.LoginBtn = widget.NewButton("login", wlogin.loginBtnCb)
	wlogin.PlayerLabel = widget.NewLabel("10001")

	row1 := container.NewBorder(nil, nil, widget.NewLabel("host:"), nil, wlogin.HostSelect)
	row2 := container.NewBorder(nil, nil, widget.NewLabel("account:"), nil, wlogin.AccountInput)
	row3 := container.NewBorder(nil, nil, widget.NewLabel("serverId:"), nil, wlogin.ServerSelect)
	row4 := wlogin.LoginBtn
	row5 := container.NewBorder(nil, nil, widget.NewLabel("player:"), nil, wlogin.PlayerLabel)

	wlogin.Container = container.NewCenter(container.NewVBox(row1, row2, row3, row4, row5))

	wlogin.ServerSelect.SetText("1")
	wlogin.AccountInput.SetText("acc_test")
	if idx := slices.Index(hosts, "local"); idx >= 0 {
		wlogin.HostSelect.SetSelectedIndex(idx)
		wlogin.currentHost = "local"
	}
	cli.SetDisconnectCb(func(err error) {
		if wlogin.status == status_DisConnect {
			return
		}
		wlogin.status = status_DisConnect
		wlogin.LoginBtn.Enable()
		wlogin.LoginBtn.SetText("login")
		wlogin.PlayerLabel.SetText("offline")
		str := fmt.Sprintf("network disconnect: %v", err)
		dialog.ShowInformation("", str, p)
	})

	return wlogin
}

func (w *LoginWidget) hostSelected(s string) {
	if w.currentHost == s {
		return
	}
	//重置
	w.currentHost = s
	if w.status == status_Connected {
		w.client.Stop()
	}
	w.status = status_DisConnect
	w.LoginBtn.Enable()
	w.LoginBtn.SetText("login")
	w.PlayerLabel.SetText("")
}

func (w *LoginWidget) loginBtnCb() {
	if w.status == status_DisConnect {
		w.status = status_Connecting
		w.LoginBtn.Disable()
		w.PlayerLabel.SetText("connecting ...")
		if err := w.loginAccount(); err != nil {
			w.status = status_DisConnect
			w.LoginBtn.Enable()
			w.PlayerLabel.SetText("offline")
			errMsg := fmt.Sprintf("login failed: %v", err)
			dialog.NewInformation("", errMsg, w.parent).Show()
		} else {
			w.status = status_Connected
			w.LoginBtn.Enable()
			w.LoginBtn.SetText("logout")
			w.PlayerLabel.SetText(strconv.Itoa(int(w.client.GetPlayerId())))
			if afterLogin != nil {
				afterLogin(w)
			}
		}
	} else if w.status == status_Connected {
		w.status = status_DisConnect
		w.LoginBtn.Enable()
		w.LoginBtn.SetText("login")
		w.PlayerLabel.SetText("offline")
		w.client.Stop()
	}
}

func (w *LoginWidget) loginAccount() error {
	//账号其他操作
	return w.connectServer()
}

func (w *LoginWidget) connectServer() error {
	if acc := w.AccountInput.Text; acc != "" {
		w.client.SetAccount(acc)
	}
	serverId := w.ServerSelect.Number()
	w.client.SetServerId(serverId)

	err := w.client.ConnectHost(w.currentHost)
	if err != nil {
		fmt.Print(err)
		return err
	}
	return w.reqLogin()
}

func (w *LoginWidget) reqLogin() error {
	if w.client.GetAccount() == "" {
		return fmt.Errorf("account is empty")
	}
	req := &hall.CLogin{
		Account: w.client.GetAccount(),
	}
	ok := w.client.SendMsg(req)
	if !ok {
		return fmt.Errorf("send login failed")
	}
	return nil
}
