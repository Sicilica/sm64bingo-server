package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/sicilica/slogging"
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

	ctx := context.Background()
	ctx = bingo.WithLogger(ctx, slog.New(
		slogging.NewPrettyHandler(os.Stdout, nil),
	))

	log.Fatal(listen(ctx, *port))
}

func listen(ctx context.Context, port int) error {
	bingo.Logger(ctx).Info("listening", slog.Int("port", port))

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

		go handleConn(ctx, cli)
	}
}

func handleConn(ctx context.Context, conn *net.TCPConn) {
	logger := bingo.Logger(ctx).With("client", conn.RemoteAddr().String())
	ctx = bingo.WithLogger(ctx, logger)

	var room bingo.Room
	var slot int

	defer conn.Close()
	defer func() {
		if room != nil {
			room.RemovePlayer(ctx, conn)
		}
	}()

	err := func() error {
		for {
			if err := conn.SetReadDeadline(time.Now().Add(IDLE_TIMEOUT)); err != nil {
				return err
			}
			h, err := message.ReadHeader(conn)
			if err != nil {
				if err == io.EOF {
					return nil
				}
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

				if hello.Name == "" {
					return errors.New("hello without name")
				}

				logger = logger.With(slog.String("name", hello.Name))
				ctx = bingo.WithLogger(ctx, logger)

				room = globalRoom
				slot, err = room.AddPlayer(ctx, conn, hello.Name, hello.PreferredColor)
				if err != nil {
					return err
				}

				logger = logger.With(slog.Int("slot", slot))
				ctx = bingo.WithLogger(ctx, logger)
			} else {
				if err := room.HandlePlayerRequest(ctx, slot, m); err != nil {
					return err
				}
			}
		}
	}()
	if err != nil {
		logger.ErrorContext(ctx, "connection error", slog.String("error", err.Error()))
	}
}
