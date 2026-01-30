package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type MessageDialog struct {
	Container *fyne.Container
	LeftText  *widget.Entry
	RightText *widget.Entry
	Window    fyne.Window
	parent    fyne.Window
}

func NewMessageDialog(p fyne.Window, left, right string) *MessageDialog {
	w := &MessageDialog{
		parent: p,
	}
	if len(left) == 0 {
		left = "{}"
	}
	if len(right) == 0 {
		right = "{}"
	}
	w.LeftText = widget.NewMultiLineEntry()
	w.LeftText.SetText(left)
	w.RightText = widget.NewMultiLineEntry()
	w.RightText.SetText(right)
	box := container.NewHSplit(w.LeftText, w.RightText)
	w.Container = container.NewStack(box)
	return w
}

func ShowMessageDislog(app fyne.App, master fyne.Window, left, right string, closeFunc func()) *MessageDialog {
	psize := master.Canvas().Size()
	size := fyne.Size{
		Width:  psize.Width / 2,
		Height: psize.Height / 2,
	}
	dialog := NewMessageDialog(master, left, right)
	box := dialog.Container
	ww := app.NewWindow("message dialog")
	dialog.Window = ww
	ww.SetContent(box)
	ww.Resize(size)
	ww.SetOnClosed(closeFunc)
	ww.Show()
	return dialog
}
