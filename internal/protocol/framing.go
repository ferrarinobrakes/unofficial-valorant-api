package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	v1 "github.com/ferrarinobrakes/unofficial-valorant-api/gen"
	"google.golang.org/protobuf/proto"
)

func ReadMessage(conn net.Conn) (*v1.Message, error) {
	var length uint32
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return nil, fmt.Errorf("failed to read length: %w", err)
	}

	if length == 0 || length > 10*1024*1024 { // Max 10MB
		return nil, fmt.Errorf("invalid message length: %d", length)
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	var msg v1.Message
	if err := proto.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}

func WriteMessage(conn net.Conn, msg *v1.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	length := uint32(len(data))
	if err := binary.Write(conn, binary.BigEndian, length); err != nil {
		return fmt.Errorf("failed to write length: %w", err)
	}
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}
