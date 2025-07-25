package log

import (
	"encoding/json"
	"errors"
	"telescope/config"
)

type Serializer interface {
	Marshal(Entry) ([]byte, error)
	Unmarshal([]byte) (Entry, error)
	Version() uint64
}

func GetSerializer(version uint64) (Serializer, error) {
	switch version {
	case config.HUMAN_READABLE_SERIALIZER:
		return humanReadableSerializer{}, nil
	case config.BINARY_SERIALIZER:
		return binarySerializer{}, nil
	default:
		return nil, errors.New("serializer not found")
	}
}

type humanReadableSerializer struct{}

func (humanReadableSerializer) Marshal(e Entry) ([]byte, error) {
	b, err := json.Marshal(e)
	// padding for human readability
	b1 := []byte{' '}
	b1 = append(b1, b...)
	b1 = append(b1, '\n')

	return b1, err
}

func (humanReadableSerializer) Unmarshal(b []byte) (e Entry, err error) {
	err = json.Unmarshal(b, &e)
	return e, err
}

func (humanReadableSerializer) Version() uint64 {
	return config.HUMAN_READABLE_SERIALIZER
}

type binarySerializer struct{}

var commandToByte = map[Command]byte{
	CommandSetVersion: 0,
	CommandType:       1,
	CommandEnter:      2,
	CommandBackspace:  3,
	CommandDelete:     4,
}
var byteToCommand = map[byte]Command{
	0: CommandSetVersion,
	1: CommandType,
	2: CommandEnter,
	3: CommandBackspace,
	4: CommandDelete,
}

func consume(buffer []byte, n int) ([]byte, []byte) {
	if len(buffer) < n {
		return buffer, make([]byte, n)
	}
	return buffer[n:], buffer[:n]
}

func (binarySerializer) Marshal(e Entry) ([]byte, error) {
	var buffer []byte
	buffer = append(buffer, commandToByte[e.Command])
	switch e.Command {
	case CommandSetVersion:
		buffer = append(buffer, uint64ToBytes(e.Version)...)
		return buffer, nil
	case CommandType:
		buffer = append(buffer, uint64ToBytes(e.CursorRow)...)
		buffer = append(buffer, uint64ToBytes(e.CursorCol)...)
		buffer = append(buffer, runeToBytes(e.Rune)...)
		return buffer, nil
	case CommandEnter, CommandBackspace, CommandDelete:
		buffer = append(buffer, uint64ToBytes(e.CursorRow)...)
		buffer = append(buffer, uint64ToBytes(e.CursorCol)...)
		return buffer, nil
	}
	return nil, errors.New("command not found")
}

func (binarySerializer) Unmarshal(buffer []byte) (e Entry, err error) {
	buffer, b := consume(buffer, 1)
	e.Command = byteToCommand[b[0]]
	switch e.Command {
	case CommandSetVersion:
		buffer, b = consume(buffer, 8)
		e.Version = bytesToUint64(b)
		return e, nil
	case CommandType:
		buffer, b = consume(buffer, 8)
		e.CursorRow = bytesToUint64(b)
		buffer, b = consume(buffer, 8)
		e.CursorCol = bytesToUint64(b)
		buffer, b = consume(buffer, 4)
		e.Rune = bytesToRune(b)
		return e, nil
	case CommandEnter, CommandBackspace, CommandDelete:
		buffer, b = consume(buffer, 8)
		e.CursorRow = bytesToUint64(b)
		buffer, b = consume(buffer, 8)
		e.CursorCol = bytesToUint64(b)
		return e, nil
	}
	return e, errors.New("parse error")
}

func (binarySerializer) Version() uint64 {
	return config.BINARY_SERIALIZER
}
