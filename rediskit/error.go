package rediskit

import (
	"errors"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrNotFound indicates that a key does not exist
	ErrNotFound = errors.New("key does not exist")

	// ErrUnsupportedEncoding indicates an unsupported encoding type was specified
	ErrUnsupportedEncoding = errors.New("unsupported encoding type")

	// ErrInvalidProtobufMessage indicates that the provided message does not implement proto.Message
	ErrInvalidProtobufMessage = errors.New("invalid protobuf message")
)

// HandleError processes Redis errors and returns appropriate custom errors
func HandleError(err error) error {
	if errors.Is(err, redis.Nil) {
		return ErrNotFound
	}
	// Add more error handling as needed
	return err
}
