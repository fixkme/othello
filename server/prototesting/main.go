package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"github.com/fixkme/othello/server/prototesting/net"
	"github.com/fixkme/othello/server/prototesting/pb"
	mtheme "github.com/fixkme/othello/server/prototesting/theme"
	"github.com/fixkme/othello/server/prototesting/ui"
)

func main() {
	fyneWindow()
}

func fyneWindow() {

	pb.RegisterMessage()

	app := app.New()
	var customTheme = &mtheme.CustomTheme{Theme: theme.DefaultTheme()}
	app.Settings().SetTheme(customTheme)

	window := app.NewWindow("protoTesting")
	window.Resize(fyne.Size{Width: 1000, Height: 800})

	cli := net.NewClient()
	wlogin := ui.NewLoginWidget(window, cli)
	wproto := ui.NewProtoWidget(window, cli)
	wmsg := ui.NewMessageWidget(app, window, cli.GetMsgReader())

	vbox := container.NewVSplit(wlogin.Container, wproto.Container)
	hbox := container.NewHSplit(vbox, wmsg.Container)
	window.SetContent(hbox)
	window.ShowAndRun()
}
