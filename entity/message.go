package entity

import (
	b "bytes"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const (
	nullByte = "\x00"
)

var (
	ErrBadMessage = errors.New("ErrorBadMessage")
)

type ChatMessage struct {
	Time    time.Time
	Content string
	Author  string
}

type DiscoveryMessage struct {
	Id     string   `json:"id"`
	Name   string   `json:"name"`
	PubKey *big.Int `json:"pub_key"`
	Port   string   `json:"port"`
}

func (d *DiscoveryMessage) ToBytes() []byte {
	msg := fmt.Sprintf("P2P|%s|%s|%s|%s", d.Id, d.Name, d.PubKey, d.Port)
	return []byte(msg)
}

func DiscoveryMessageFromBytes(bytes []byte, fcP2P, fcGe func([]string) error) error {
	bytes = b.Trim(bytes, nullByte)
	arrayStr := strings.Split(string(bytes), "|")
	if len(arrayStr) != 5 {
		return ErrBadMessage
	}
	if arrayStr[0] == "P2P" {
		return fcP2P(arrayStr[1:])
	} else {
		return fcGe(arrayStr)
	}
}
