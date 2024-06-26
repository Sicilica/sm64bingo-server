package bingo

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"math/rand/v2"
	"net"
	"slices"
	"sync"

	"github.com/sicilica/sm64bingo-server/message"
)

type Room interface {
	AddPlayer(context.Context, net.Conn, string, int) (int, error)
	RemovePlayer(context.Context, net.Conn)
	HandlePlayerRequest(context.Context, int, any) error
}

type room struct {
	sync.Mutex
	Board      message.Board
	Colors     []uint32
	NumPlayers int
	Players    []*playerInfo
}

var _ Room = &room{}

func NewRoom(maxPlayers int, colors []uint32) Room {
	return &room{
		Colors:  colors,
		Players: make([]*playerInfo, maxPlayers),
	}
}

func (r *room) AddPlayer(ctx context.Context, conn net.Conn, name string, prefColor int) (int, error) {
	r.Lock()
	defer r.Unlock()

	slot, err := r.findEmptySlot()
	if err != nil {
		return 0, err
	}

	r.Players[slot] = &playerInfo{
		Name:  name,
		Color: r.findUnusedColor(prefColor),
		Conn:  conn,
	}
	r.NumPlayers += 1
	Logger(ctx).Info("connected", slog.Int("slot", slot))

	// When the first player joins an empty room, reset the board...?
	if r.NumPlayers == 1 {
		r.Board = message.Board{}
	}

	// Send new player's state to all clients
	r.broadcast(&message.PlayerConnected{
		Slot:  slot,
		Name:  r.Players[slot].Name,
		Color: r.Players[slot].Color,
	})

	// Send the board and all other players' states to the new client
	if err := func() error {
		if err := message.Write(conn, &r.Board); err != nil {
			return err
		}

		for i, p := range r.Players {
			if p == nil || i == slot {
				continue
			}
			if err := message.Write(conn, &message.PlayerConnected{
				Slot:  i,
				Name:  p.Name,
				Color: p.Color,
			}); err != nil {
				return err
			}
			if err := message.Write(conn, &message.PlayerLocation{
				Slot:     i,
				Location: p.Location,
			}); err != nil {
				return err
			}
			if err := message.Write(conn, &message.PlayerCompletion{
				Slot:       i,
				Completion: p.Completion,
			}); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		log.Println(err)
	}

	return slot, nil
}

func (r *room) RemovePlayer(ctx context.Context, conn net.Conn) {
	r.Lock()
	defer r.Unlock()

	slot := slices.IndexFunc(r.Players, func(p *playerInfo) bool {
		return p != nil && p.Conn == conn
	})
	if slot < 0 {
		return
	}
	Logger(ctx).Info("disconnected", slog.Int("slot", slot))

	r.Players[slot] = nil
	r.NumPlayers -= 1

	r.broadcast(&message.PlayerDisconnected{
		Slot: slot,
	})
}

func (r *room) HandlePlayerRequest(ctx context.Context, slot int, m any) error {
	if _, ok := m.(*message.Hello); ok {
		return nil
	} else if req, ok := m.(*message.Board); ok {
		return r.GenerateNewBoard(ctx, req.Config)
	} else if req, ok := m.(*message.PlayerCompletion); ok {
		return r.SetPlayerCompletion(ctx, slot, req.Completion)
	} else if req, ok := m.(*message.PlayerLocation); ok {
		return r.SetPlayerLocation(ctx, slot, req.Location)
	}
	return errors.New("unhandled player request")
}

func (r *room) GenerateNewBoard(ctx context.Context, cfg string) error {
	r.Lock()
	defer r.Unlock()

	r.Board = message.Board{
		Seed:   rand.Int32(),
		Config: cfg,
	}
	Logger(ctx).Info("generated new board", slog.Int("seed", int(r.Board.Seed)), slog.String("config", r.Board.Config))

	r.broadcast(&r.Board)
	return nil
}

func (r *room) SetPlayerCompletion(ctx context.Context, slot int, v uint32) error {
	r.Lock()
	defer r.Unlock()

	if r.Players[slot] == nil {
		return errors.New("slot is empty")
	}

	Logger(ctx).Debug("update completion", slog.Int("completion", int(v)))

	r.Players[slot].Completion = v
	r.broadcast(&message.PlayerCompletion{
		Slot:       slot,
		Completion: v,
	})
	return nil
}

func (r *room) SetPlayerLocation(ctx context.Context, slot int, v int16) error {
	r.Lock()
	defer r.Unlock()

	if r.Players[slot] == nil {
		return errors.New("slot is empty")
	}

	Logger(ctx).Debug("update location", slog.Int("location", int(v)))

	r.Players[slot].Location = v
	r.broadcast(&message.PlayerLocation{
		Slot:     slot,
		Location: v,
	})
	return nil
}

func (r *room) broadcast(msg any) {
	for _, p := range r.Players {
		if p == nil {
			continue
		}
		if err := message.Write(p.Conn, msg); err != nil {
			log.Println(err)
		}
	}
}

func (r *room) findEmptySlot() (int, error) {
	for i, p := range r.Players {
		if p == nil {
			return i, nil
		}
	}
	return 0, errors.New("no available slots")
}

func (r *room) findUnusedColor(prefColor int) uint32 {
	if prefColor >= 0 && prefColor < len(r.Colors) {
		if r.isColorUnused(r.Colors[prefColor]) {
			return r.Colors[prefColor]
		}
	}
	for i, color := range r.Colors {
		if i == prefColor {
			continue
		}
		if r.isColorUnused(color) {
			return color
		}
	}
	return r.Colors[0]
}

func (r *room) isColorUnused(color uint32) bool {
	for _, p := range r.Players {
		if p != nil && p.Color == color {
			return false
		}
	}
	return true
}
