package ui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
)

type NumericalEntry struct {
	widget.Entry
}

func NewNumericalEntry() *NumericalEntry {
	entry := &NumericalEntry{}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (e *NumericalEntry) TypedKey(key *fyne.KeyEvent) {
	e.Entry.TypedKey(key)
}

func (e *NumericalEntry) TypedRune(r rune) {
	if r >= '0' && r <= '9' {
		e.Entry.TypedRune(r)
	}
}

func (e *NumericalEntry) TypedShortcut(shortcut fyne.Shortcut) {
	switch v := shortcut.(type) {
	case *fyne.ShortcutPaste:
		content := v.Clipboard.Content()
		if _, err := strconv.ParseInt(content, 10, 64); err == nil {
			e.Entry.TypedShortcut(shortcut)
		}
	default:
		e.Entry.TypedShortcut(shortcut)
	}
}

func (e *NumericalEntry) Keyboard() mobile.KeyboardType {
	return mobile.NumberKeyboard
}

func (e *NumericalEntry) Number() int64 {
	if len(e.Text) == 0 {
		return 0
	}
	val, _ := strconv.ParseInt(e.Text, 10, 64)
	return val
}
