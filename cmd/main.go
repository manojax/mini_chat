package main

import (
	"context"
	"log"

	"chat_tool/ui"
)

const (
	Version             = "1.0.0-Dev"
	BroadcastChanBuffer = 10
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	if err := ui.NewApp(BroadcastChanBuffer).Run(ctx, Version); err != nil {
		log.Fatal(err)
	}
	cancel()
}
