package protocol

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

const Version = 1

type Message struct {
	Type    string `json:"type"`
	Version int    `json:"version,omitempty"`
	Name    string `json:"name,omitempty"`
	Size    int64  `json:"size,omitempty"`
	MIME    string `json:"mime,omitempty"`
	Code    string `json:"code,omitempty"`
}

func Write(w io.Writer, message Message) error {
	b, err := json.Marshal(message)
	if err != nil {
		return err
	}
	if uint64(len(b)) > uint64(^uint32(0)) {
		return fmt.Errorf("message too large")
	}
	var n [4]byte
	binary.BigEndian.PutUint32(n[:], uint32(len(b)))
	if _, err = w.Write(n[:]); err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func Read(r io.Reader) (Message, error) {
	var n [4]byte
	if _, err := io.ReadFull(r, n[:]); err != nil {
		return Message{}, err
	}
	b := make([]byte, binary.BigEndian.Uint32(n[:]))
	if _, err := io.ReadFull(r, b); err != nil {
		return Message{}, err
	}
	var message Message
	if err := json.Unmarshal(b, &message); err != nil {
		return Message{}, err
	}
	return message, nil
}
