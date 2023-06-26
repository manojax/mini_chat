package entity

import (
	"chat_tool/utils"

	"github.com/google/uuid"
)

type Owner struct {
	Id   string
	Name string
	Port string
	DH   utils.DiffieHellman
	Repo *RoomRepository
}

func NewOwner(name, port string, broadcastChanBuffer int) *Owner {
	o := &Owner{
		Id:   uuid.NewString(),
		Name: name,
		Port: port,
		DH:   utils.NewDiffieHellman(),
		Repo: NewRoomRepository(),
	}
	o.Repo.Add(&Room{
		Id:            "00000000-0000-0000-0000-00000000000",
		Name:          "General",
		Host:          "",
		Messages:      make([]*ChatMessage, 0),
		IsGeneral:     true,
		BroadcastChan: make(chan *ChatMessage, broadcastChanBuffer),
	})
	return o
}
