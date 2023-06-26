package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TextInput struct {
	View *tview.InputField
}

func NewTextInput() *TextInput {
	inputField := tview.NewInputField().
		SetPlaceholder("Type a new message").
		SetDoneFunc(func(key tcell.Key) {})
	inputField.SetBorder(true)
	return &TextInput{
		View: inputField,
	}
}
