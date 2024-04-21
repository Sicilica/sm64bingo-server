package message

import (
	"encoding/binary"
	"io"
)

type Header struct {
	Size uint16
	Type uint16
}

func ReadHeader(r io.Reader) (Header, error) {
	var b [2]uint16
	if err := binary.Read(r, binary.BigEndian, b[:]); err != nil {
		return Header{}, err
	}

	return Header{
		Type: b[0],
		Size: b[1],
	}, nil
}

func (h Header) Write(w io.Writer) error {
	var b [2]uint16
	b[0] = h.Type
	b[1] = h.Size
	return binary.Write(w, binary.BigEndian, b[:])
}
