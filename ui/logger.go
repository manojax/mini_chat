package ui

import (
	"github.com/rivo/tview"
)

type LoggerView struct {
	View *tview.TextView
}

func NewLoggerView() *LoggerView {
	messages := tview.NewTextView().
		SetText("").
		SetDynamicColors(true).
		SetScrollable(true)
	messages.SetBorder(true).SetTitle("Logs")
	return &LoggerView{
		View: messages,
	}
}

func (c *LoggerView) RenderMessages(msg string) {
	c.View.SetText(c.View.GetText(false) + msg)
}
