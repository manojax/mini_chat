package ui

import (
	"fmt"

	"github.com/rivo/tview"

	"chat_tool/entity"
)

type Sidebar struct {
	View         *tview.List
	repo         *entity.RoomRepository
	currentCount int
}

func NewSidebar(repo *entity.RoomRepository) *Sidebar {
	view := tview.NewList()
	view.SetTitle("Teams").
		SetBorder(true)
	view.SetWrapAround(true)
	return &Sidebar{
		View:         view,
		repo:         repo,
		currentCount: -1,
	}
}

func (s *Sidebar) Reprint() {
	count := len(s.repo.GetRooms())
	if s.currentCount == count {
		return
	}
	s.currentCount = count
	s.View.Clear()
	for _, room := range s.repo.GetRooms() {
		mainText := fmt.Sprintf("%s (Addr: %s)", room.Name, room.Host)
		if room.IsGeneral {
			mainText = room.Name
		}
		s.View.AddItem(mainText, room.Id, 0, nil)
	}
}
