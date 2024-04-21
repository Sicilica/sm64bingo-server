package main

import (
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"time"

	bingo "github.com/sicilica/sm64bingo-server"
	"github.com/sicilica/sm64bingo-server/message"
)

const IDLE_TIMEOUT = 30 * time.Second
const BODY_TIMEOUT = 5 * time.Second

var COLORS = []uint32{
	colorFromRGB(255, 0, 0),   // RED
	colorFromRGB(0, 0, 255),   // BLUE
	colorFromRGB(0, 255, 0),   // GREEN
	colorFromRGB(96, 0, 160),  // PURPLE
	colorFromRGB(255, 127, 0), // ORANGE
	colorFromRGB(255, 0, 255), // PINK
	colorFromRGB(0, 160, 160), // TEAL
	colorFromRGB(255, 255, 0), // YELLOW
}

// TODO: For now, there's only one room; in the future, the HELLO request should
// optionally be able to specify a room number
var globalRoom = bingo.NewRoom(8, COLORS)

func main() {
	port := flag.Int("port", 0, "port to listen on")
	flag.Parse()

	if port == nil || *port == 0 {
		flag.Usage()
		os.Exit(1)
	}

	log.Fatal(listen(*port, handleConn))
}

func listen(port int, fn func(*net.TCPConn) error) error {
	lis, err := net.ListenTCP("tcp", &net.TCPAddr{
		Port: port,
	})
	if err != nil {
		return err
	}
	defer lis.Close()

	for {
		cli, err := lis.AcceptTCP()
		if err != nil {
			return err
		}
		go func() {
			defer cli.Close()
			err := fn(cli)
			if err != nil {
				log.Printf("%s: %v\n", cli.RemoteAddr().String(), err)
			}
		}()
	}
}

func handleConn(conn *net.TCPConn) error {
	var room bingo.Room
	var slot int

	for {
		if err := conn.SetReadDeadline(time.Now().Add(IDLE_TIMEOUT)); err != nil {
			return err
		}
		h, err := message.ReadHeader(conn)
		if err != nil {
			return err
		}

		if err := conn.SetReadDeadline(time.Now().Add(BODY_TIMEOUT)); err != nil {
			return err
		}
		m, err := message.ReadBody(conn, h)
		if err != nil {
			return err
		}

		if room == nil {
			hello, ok := m.(*message.Hello)
			if !ok {
				return errors.New("first message on connection must be hello")
			}

			room = globalRoom
			slot, err = room.AddPlayer(conn, hello.Name, hello.PreferredColor)
			if err != nil {
				return err
			}
			defer room.RemovePlayer(conn)
		} else {
			if err := room.HandlePlayerRequest(slot, m); err != nil {
				return err
			}
		}
	}
}
