package connection

import (
	"context"
	"fmt"
	"net"
	"time"

	"chat_tool/entity"
	"chat_tool/utils"
)

const (
	Frequency  = 1 * time.Second
	bufferSize = 8192
)

type Broker struct {
	owner     *entity.Owner
	p2p       *P2PChannel
	broadcast *BroadcastChannel
}

func NewBroker(o *entity.Owner, broadcastIP string) *Broker {
	broadcastAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", broadcastIP, o.Port))
	if err != nil {
		utils.LL.Error("Broker: %s", err.Error())
		return nil
	}
	utils.LL.Info("Broker: ResolveUDPAddr [yellow]%s:%s[white]", broadcastIP, o.Port)
	return &Broker{
		owner:     o,
		p2p:       NewP2PChannel(fmt.Sprintf("0.0.0.0:%s", o.Port), o),
		broadcast: NewBroadcastChannel(broadcastAddr, Frequency, o),
	}
}

func (m *Broker) Start(ctx context.Context) {
	utils.LL.Info("Broker: START")
	go m.p2p.Start(ctx)
	go m.broadcast.Start(ctx)
}
