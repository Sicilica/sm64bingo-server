package message

import (
	"encoding/json"
	"fmt"
	"io"
)

func ReadBody(r io.Reader, h Header) (any, error) {
	b := make([]byte, h.Size)
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return nil, err
	}

	switch h.Type {
	case TYPE_HELLO:
		msg := &Hello{}
		return msg, json.Unmarshal(b, msg)
	case TYPE_PLAYER_CONNECTED:
		msg := &PlayerConnected{}
		return msg, json.Unmarshal(b, msg)
	case TYPE_PLAYER_DISCONNECTED:
		msg := &PlayerDisconnected{}
		return msg, json.Unmarshal(b, msg)
	case TYPE_BOARD:
		msg := &Board{}
		return msg, json.Unmarshal(b, msg)
	case TYPE_PLAYER_LOCATION:
		msg := &PlayerLocation{}
		return msg, json.Unmarshal(b, msg)
	case TYPE_PLAYER_COMPLETION:
		msg := &PlayerCompletion{}
		return msg, json.Unmarshal(b, msg)
	default:
		return nil, fmt.Errorf("unknown message type %d", h.Type)
	}
}
