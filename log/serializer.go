package log

import (
	"encoding/json"
	"errors"
)

type Serializer interface {
	Marshal(Entry) ([]byte, error)
	Unmarshal([]byte) (Entry, error)
}

func getSerializer(version int) (Serializer, error) {
	switch version {
	case 0:
		return serializerV0{}, nil
	default:
		return nil, errors.New("serializer not found")
	}
}

type serializerV0 struct{}

func (serializerV0) Marshal(e Entry) ([]byte, error) {
	return json.Marshal(e)
}

func (serializerV0) Unmarshal(b []byte) (e Entry, err error) {
	err = json.Unmarshal(b, &e)
	return e, err
}

type serializerV1 struct{}

func (serializerV1) Marshal(e Entry) ([]byte, error) {
	return json.Marshal(e)
}

func (serializerV1) Unmarshal(b []byte) (e Entry, err error) {
	err = json.Unmarshal(b, &e)
	return e, err
}
