package log

import (
	"encoding/json"
	"errors"
)

type Serializer interface {
	Marshal(Entry) ([]byte, error)
	Unmarshal([]byte) (Entry, error)
	Version() uint64
}

func GetSerializer(version uint64) (Serializer, error) {
	switch version {
	case 0:
		return serializerV0{}, nil
	case 1:
		return serializerV1{}, nil
	default:
		return nil, errors.New("serializer not found")
	}
}

type serializerV0 struct{}

func (serializerV0) Marshal(e Entry) ([]byte, error) {
	b, err := json.Marshal(e)
	// padding for human readability
	b1 := []byte{' '}
	b1 = append(b1, b...)
	b1 = append(b1, '\n')

	return b1, err
}

func (serializerV0) Unmarshal(b []byte) (e Entry, err error) {
	err = json.Unmarshal(b, &e)
	return e, err
}

func (serializerV0) Version() uint64 {
	return 0
}

type serializerV1 struct{}

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

func (serializerV1) Marshal(e Entry) ([]byte, error) {
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

func (serializerV1) Unmarshal(buffer []byte) (e Entry, err error) {
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

func (serializerV1) Version() uint64 {
	return 1
}
