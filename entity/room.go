package entity

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	"chat_tool/utils"

	"golang.org/x/net/websocket"
)

var (
	dialer     = net.Dialer{Timeout: 2 * time.Second}
	httpClient = http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
	}
	pingDelay       = time.Second
	ErrDisconnected = errors.New("disconnected")
)

type Room struct {
	Id            string
	Name          string
	PubKey        *big.Int
	Host          string
	Messages      []*ChatMessage
	IsGeneral     bool
	BroadcastChan chan *ChatMessage
	WSChan        chan string
	wsConn        *websocket.Conn
}

func (r *Room) Close() {
	if r.WSChan != nil {
		close(r.WSChan)
	}
}

func (r *Room) HandleWS(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-r.WSChan:
			if !ok {
				return
			}
			if r.wsConn != nil {
				if _, err := r.wsConn.Write([]byte(msg)); err != nil {
					utils.LL.Error("Room-HandleWS: %s", err.Error())
					return
				}
			} else {
				origin := fmt.Sprintf("http://%s/", r.Host)
				url := fmt.Sprintf("ws://%s/ws", r.Host)
				ws, err := websocket.Dial(url, "", origin)
				if err != nil {
					utils.LL.Error("Room-HandleWS: %s", err.Error())
					return
				}
				if _, err := ws.Write([]byte(msg)); err != nil {
					utils.LL.Error("Room-HandleWS: %s", err.Error())
					return
				}
				r.wsConn = ws
			}
		}
	}
}

func (r *Room) AddMessage(text, author string) {
	r.Messages = append(r.Messages, &ChatMessage{
		Time:    time.Now(),
		Content: text,
		Author:  author,
	})
}

func (r *Room) sendBroadcastMessage(id string) error {
	if len(r.Messages) > 0 {
		r.BroadcastChan <- r.Messages[len(r.Messages)-1]
	}
	return nil
}

func (r *Room) sendWSMessage(id, encryptedMessage string) error {
	r.WSChan <- fmt.Sprintf("%s|%s", id, encryptedMessage)
	return nil
}

func (r *Room) SendMessage(id, message string, dh utils.DiffieHellman) error {
	if r.IsGeneral {
		return r.sendBroadcastMessage(id)
	}

	encryptedMessage, err := utils.EncryptMessage(utils.GetSecret(r.PubKey, dh), message)
	if err != nil {
		return err
	}
	return r.sendWSMessage(id, encryptedMessage)
}

type RoomRepository struct {
	rwMutex *sync.RWMutex
	rooms   map[string]*Room
	Updated chan string
}

func NewRoomRepository() *RoomRepository {
	repo := &RoomRepository{
		rwMutex: &sync.RWMutex{},
		rooms:   make(map[string]*Room),
		Updated: make(chan string),
	}
	repo.Ping()
	return repo
}

func (r *RoomRepository) Add(room *Room) {
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()
	if _, found := r.rooms[room.Id]; !found {
		r.rooms[room.Id] = room
	}
}

func (r *RoomRepository) Delete(id string) {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()
	delete(r.rooms, id)
}

func (r *RoomRepository) Get(id string) (*Room, bool) {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()
	room, found := r.rooms[id]
	return room, found
}

func (r *RoomRepository) GetGeneralRooms() []*Room {
	roomsSlice := make([]*Room, 0)
	for _, room := range r.rooms {
		if room.IsGeneral {
			roomsSlice = append(roomsSlice, room)
		}
	}
	return roomsSlice
}

func (r *RoomRepository) GetRooms() []*Room {
	roomsSlice := make([]*Room, 0, len(r.rooms))
	for _, room := range r.rooms {
		roomsSlice = append(roomsSlice, room)
	}
	sort.Slice(roomsSlice, func(i, j int) bool {
		return roomsSlice[i].Id < roomsSlice[j].Id
	})
	return roomsSlice
}

func (r *RoomRepository) Ping() {
	ticker := time.NewTicker(pingDelay)

	go func() {
		for {
			<-ticker.C
			for _, room := range r.GetRooms() {
				if room.IsGeneral {
					continue
				}
				req, err := http.NewRequest("HEAD", "http://"+room.Host, nil)
				if err != nil {
					utils.LL.Error("PING: NewRequest %s", err.Error())
					continue
				}
				resp, err := httpClient.Do(req)
				if err != nil {
					utils.LL.Error("PING: Do %s", err.Error())
					room.Close()
					r.Delete(room.Id)
					r.Updated <- room.Id
					continue
				}
				resp.Body.Close()
				if resp.StatusCode != 200 {
					utils.LL.Error("PING: Status %d", resp.StatusCode)
				}
			}
		}
	}()
}
