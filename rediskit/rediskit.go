package rediskit

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
)

// ExampleProto is a sample Protobuf message.
// In a real-world scenario, replace this with your actual Protobuf-generated structs.
type ExampleProto struct {
	Id    int32  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name  string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Email string `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
}

func (m *ExampleProto) Reset()         { *m = ExampleProto{} }
func (m *ExampleProto) String() string { return proto.CompactTextString(m) }
func (*ExampleProto) ProtoMessage()    {}

// CacheProto caches a Protobuf message
func (rk *RedisKitClient) CacheProto(ctx context.Context, key string, data proto.Message, expiration time.Duration) error {
	return rk.Set(ctx, key, data, expiration)
}

// GetProto retrieves a Protobuf message from cache
func (rk *RedisKitClient) GetProto(ctx context.Context, key string, dest proto.Message) error {
	return rk.Get(ctx, key, dest)
}

// Example usage functions

// CacheExampleStruct caches an ExampleStruct using the selected encoder
func (rk *RedisKitClient) CacheExampleStruct(ctx context.Context, key string, data interface{}, expiration time.Duration) error {
	return rk.Set(ctx, key, data, expiration)
}

// GetExampleStruct retrieves an ExampleStruct from cache
func (rk *RedisKitClient) GetExampleStruct(ctx context.Context, key string, dest interface{}) error {
	return rk.Get(ctx, key, dest)
}
