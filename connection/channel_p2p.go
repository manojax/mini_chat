package connection

import (
	"context"
	"net/http"
	"strings"
	"time"

	"chat_tool/entity"
	"chat_tool/utils"

	"golang.org/x/net/websocket"
)

type P2PChannel struct {
	addr  string
	owner *entity.Owner
}

func NewP2PChannel(addr string, o *entity.Owner) *P2PChannel {
	return &P2PChannel{
		owner: o,
		addr:  addr,
	}
}

func (d *P2PChannel) Start(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	mux.Handle("/ws", websocket.Handler(func(c *websocket.Conn) {
		utils.LL.Info("WS: Handshake")
		var msg = make([]byte, bufferSize)
		var n int
		var err error
		for {
			if n, err = c.Read(msg); err != nil {
				utils.LL.Error("WS: Read, %s", err.Error())
				break
			}

			arr := strings.Split(string(msg[:n]), "|")
			roomId := arr[0]
			messageText := arr[1]
			peer, found := d.owner.Repo.Get(roomId)
			if !found {
				continue
			}

			decryptedMessage, err := utils.DecryptMessage(utils.GetSecret(peer.PubKey, d.owner.DH), messageText)
			if err != nil {
				utils.LL.Error("WS: CHAT %s", err.Error())
				continue
			}
			peer.AddMessage(decryptedMessage, peer.Name)
		}
		utils.LL.Info("WS: END")
	}))

	server := &http.Server{
		Addr:    d.addr,
		Handler: mux,
	}

	go func() {
		utils.LL.Info("P2P: ListenAndServe %s", d.addr)
		if err := server.ListenAndServe(); err != nil {
			utils.LL.Error("P2P: ListenAndServe %s", err.Error())
		}
	}()

	<-ctx.Done()
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctxTimeout)
	utils.LL.Info("P2P: Shutdown")
}
