package message

import (
	"encoding/json"
	"fmt"
	"io"
)

func Write(w io.Writer, m any) error {
	if _, ok := m.(*Hello); ok {
		return write(w, TYPE_HELLO, m)
	} else if _, ok := m.(*PlayerConnected); ok {
		return write(w, TYPE_PLAYER_CONNECTED, m)
	} else if _, ok := m.(*PlayerDisconnected); ok {
		return write(w, TYPE_PLAYER_DISCONNECTED, m)
	} else if _, ok := m.(*Board); ok {
		return write(w, TYPE_BOARD, m)
	} else if _, ok := m.(*PlayerLocation); ok {
		return write(w, TYPE_PLAYER_LOCATION, m)
	} else if _, ok := m.(*PlayerCompletion); ok {
		return write(w, TYPE_PLAYER_COMPLETION, m)
	}
	return fmt.Errorf("unhandled message type")
}

func write(w io.Writer, t uint16, m any) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	h := Header{
		Type: t,
		Size: uint16(len(b)),
	}
	if err := h.Write(w); err != nil {
		return err
	}

	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}
