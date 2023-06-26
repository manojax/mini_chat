package connection

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"time"

	"chat_tool/entity"
	"chat_tool/utils"
)

type BroadcastChannel struct {
	addr      *net.UDPAddr
	frequency time.Duration
	owner     *entity.Owner
}

func NewBroadcastChannel(addr *net.UDPAddr, frequency time.Duration, o *entity.Owner) *BroadcastChannel {
	return &BroadcastChannel{
		addr:      addr,
		frequency: frequency,
		owner:     o,
	}
}

func (d *BroadcastChannel) Start(ctx context.Context) {
	utils.LL.Info("BroadcastChannel: START")
	go d.startCasting(ctx)
	go d.listenCasting(ctx)
}

func (d *BroadcastChannel) broadcastMessage(ctx context.Context, conn *net.UDPConn, prefix string, channel chan *entity.ChatMessage) {
	if conn == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-channel:
			if !ok {
				return
			}
			_, err := conn.Write([]byte(fmt.Sprintf("%s|%s|%s|%s",
				prefix,
				msg.Time.Format(time.RFC3339),
				msg.Content,
				msg.Author,
			)))
			if err != nil {
				utils.LL.Error("BroadcastMessage: %s", err.Error())
				return
			}
		}
	}
}

func (d *BroadcastChannel) startCasting(ctx context.Context) {
	conn, err := net.DialUDP("udp", nil, d.addr)
	if err != nil {
		utils.LL.Error("BroadcastChannel: Start casting %s", err.Error())
		return
	}

	for _, r := range d.owner.Repo.GetGeneralRooms() {
		go d.broadcastMessage(ctx, conn, fmt.Sprintf("%s|%s", r.Id, d.owner.Id), r.BroadcastChan)
	}

	ticker := time.NewTicker(d.frequency)
	for {
		select {
		case <-ctx.Done():
			conn.Close()
			return
		case <-ticker.C:
			msg := &entity.DiscoveryMessage{
				Id:     d.owner.Id,
				Name:   d.owner.Name,
				PubKey: d.owner.DH.PublicKey,
				Port:   d.owner.Port,
			}
			_, err = conn.Write(msg.ToBytes())
			if err != nil {
				utils.LL.Error("BroadcastChannel: Casting %s", err.Error())
				return
			}
		}
	}
}

func (d *BroadcastChannel) listenCasting(ctx context.Context) {
	conn, err := net.ListenMulticastUDP("udp", nil, d.addr)
	if err != nil {
		utils.LL.Error("ListenMulticastUDP: %s", err.Error())
		return
	}
	err = conn.SetReadBuffer(bufferSize)
	if err != nil {
		utils.LL.Error("SetReadBuffer: %s", err.Error())
		return
	}
	for {
		select {
		case <-ctx.Done():
			conn.Close()
			return
		default:
			rawBytes, addr, err := utils.ReadFromUDPConnection(conn, bufferSize)
			if err != nil {
				utils.LL.Error("ReadFromUDPConnection: %s", err.Error())
				return
			}

			err = entity.DiscoveryMessageFromBytes(rawBytes, func(s []string) error {
				k := new(big.Int)
				k, ok := k.SetString(s[2], 10)
				if !ok {
					return fmt.Errorf("invalid pubkey")
				}
				room := &entity.Room{
					Id:        s[0],
					Name:      s[1],
					PubKey:    k,
					Host:      fmt.Sprintf("%s:%s", addr.IP.String(), s[3]),
					Messages:  make([]*entity.ChatMessage, 0),
					IsGeneral: false,
					WSChan:    make(chan string, 10),
				}
				if room.Id != d.owner.Id {
					if _, roomFound := d.owner.Repo.Get(room.Id); !roomFound {
						utils.LL.Info("ListenCasting: JOINING [green]%s[white] - [yellow]%s[white]", room.Name, room.Host)
						d.owner.Repo.Add(room)
						go room.HandleWS(ctx)
					}
				}
				return nil
			}, func(s []string) error {
				if s[1] != d.owner.Id {
					if t, err := time.Parse(time.RFC3339, s[2]); err == nil {
						if r, ok := d.owner.Repo.Get(s[0]); ok {
							utils.LL.Info("ListenCasting: MESSAGE from [green]%s[white]", s[4])
							r.Messages = append(r.Messages, &entity.ChatMessage{
								Time:    t,
								Content: s[3],
								Author:  s[4],
							})
						}
					}
				}
				return nil
			})
			if err != nil {
				utils.LL.Error("DiscoveryMessage: %s", err.Error())
				return
			}
		}
	}
}
