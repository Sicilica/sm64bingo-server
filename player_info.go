package bingo

import (
	"net"
)

type playerInfo struct {
	Color      uint32
	Conn       net.Conn
	Name       string
	Location   int16
	Completion uint32
}
