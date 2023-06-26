package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"chat_tool/entity"
)

type TextView struct {
	View                *tview.TextView
	currentMessageCount int
}

func NewTextView() *TextView {
	messages := tview.NewTextView().
		SetText("").
		SetDynamicColors(true).
		SetScrollable(true)
	messages.SetBorder(true)
	return &TextView{
		View:                messages,
		currentMessageCount: 0,
	}
}

func (c *TextView) RenderMessages(messages []*entity.ChatMessage, selfName string) {
	if c.currentMessageCount == len(messages) {
		return
	}
	c.currentMessageCount = len(messages)
	text := strings.Repeat("\n", maxMessagesInView)
	for _, message := range messages {
		text += fmt.Sprintf("%s %s: %s\n\n",
			formatTime(message),
			formatAuthor(message, message.Author == selfName),
			formatText(message))
	}
	c.View.SetText(text[:len(text)-1]).ScrollToEnd()
}

func formatTime(message *entity.ChatMessage) string {
	now := message.Time.UTC()
	return fmt.Sprintf("%s%s", "[blue]", now.Format(timeFormat))
}

func formatAuthor(message *entity.ChatMessage, isAuthor bool) string {
	if isAuthor {
		return fmt.Sprintf("%s%s", "[green]", message.Author)
	}
	return fmt.Sprintf("%s%s", "[red]", message.Author)
}

func formatText(message *entity.ChatMessage) string {
	return fmt.Sprintf("%s%s", "[white]", message.Content)
}
