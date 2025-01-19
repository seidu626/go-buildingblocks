package rediskit

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/vmihailenco/msgpack/v5"
)

// Encoder defines the methods required for encoding and decoding data
type Encoder interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

// JSONEncoder implements Encoder using JSON
type JSONEncoder struct{}

func (je *JSONEncoder) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (je *JSONEncoder) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// MsgpackEncoder implements Encoder using Msgpack
type MsgpackEncoder struct{}

func (me *MsgpackEncoder) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (me *MsgpackEncoder) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

// ProtobufEncoder implements Encoder using Protobuf
type ProtobufEncoder struct{}

func (pe *ProtobufEncoder) Marshal(v interface{}) ([]byte, error) {
	pb, ok := v.(proto.Message)
	if !ok {
		return nil, ErrInvalidProtobufMessage
	}
	return proto.Marshal(pb)
}

func (pe *ProtobufEncoder) Unmarshal(data []byte, v interface{}) error {
	pb, ok := v.(proto.Message)
	if !ok {
		return ErrInvalidProtobufMessage
	}
	return proto.Unmarshal(data, pb)
}

// SelectEncoder returns the appropriate Encoder based on the encoding type
func SelectEncoder(encoding string) (Encoder, error) {
	switch encoding {
	case "json":
		return &JSONEncoder{}, nil
	case "msgpack":
		return &MsgpackEncoder{}, nil
	case "protobuf":
		return &ProtobufEncoder{}, nil
	default:
		return nil, ErrUnsupportedEncoding
	}
}
