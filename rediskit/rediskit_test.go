package rediskit

import (
	"context"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

type MockRedisClient struct {
	mock   redismock.ClientMock
	client *redis.Client
}

func NewMockRedisClient(encoding string) (*MockRedisClient, *RedisKitClient, error) {
	client, mock := redismock.NewClientMock()
	cfg := Config{
		Addr:              "localhost:6379",
		Password:          "",
		DB:                0,
		PoolSize:          10,
		MinIdleConns:      2,
		IdleTimeout:       5 * time.Minute,
		MaxRetries:        3,
		MinRetryBackoff:   100 * time.Millisecond,
		MaxRetryBackoff:   1 * time.Second,
		DefaultExpiration: 5 * time.Minute,
		Encoding:          encoding,
		IsCluster:         false,
	}

	encoder, err := SelectEncoder(cfg.Encoding)
	if err != nil {
		return nil, nil, err
	}

	// Initialize cache with mock client
	cacheInstance := NewRedisCache(client, encoder, cfg.DefaultExpiration)

	rk := &RedisKitClient{
		client:    client,
		cache:     cacheInstance,
		config:    cfg,
		isCluster: cfg.IsCluster,
		encoder:   encoder,
	}

	return &MockRedisClient{mock: mock, client: client}, rk, nil
}

func TestSetAndGetJSON(t *testing.T) {
	mc, rk, err := NewMockRedisClient("json")
	assert.NoError(t, err)

	ctx := context.Background()
	key := "test_key"
	value := map[string]interface{}{
		"id":    1,
		"name":  "John Doe",
		"email": "john@example.com",
	}

	// Mock SET operation
	mc.mock.ExpectSet(key, `{"id":1,"name":"John Doe","email":"john@example.com"}`, 5*time.Minute).SetVal("OK")

	err = rk.Set(ctx, key, value, 5*time.Minute)
	assert.NoError(t, err)

	// Mock GET operation
	mc.mock.ExpectGet(key).SetVal(`{"id":1,"name":"John Doe","email":"john@example.com"}`)

	var result map[string]interface{}
	err = rk.Get(ctx, key, &result)
	assert.NoError(t, err)
	assert.Equal(t, value, result)

	// Ensure all expectations were met
	assert.NoError(t, mc.mock.ExpectationsWereMet())
}

func TestSetAndGetProtobuf(t *testing.T) {
	mc, rk, err := NewMockRedisClient("protobuf")
	assert.NoError(t, err)

	ctx := context.Background()
	key := "proto_key"
	value := &ExampleProto{
		Id:    1,
		Name:  "Alice",
		Email: "alice@example.com",
	}

	marshaled, err := proto.Marshal(value)
	assert.NoError(t, err)

	// Mock SET operation
	mc.mock.ExpectSet(key, marshaled, 5*time.Minute).SetVal("OK")

	err = rk.CacheProto(ctx, key, value, 5*time.Minute)
	assert.NoError(t, err)

	// Mock GET operation
	mc.mock.ExpectGet(key).SetVal(string(marshaled))

	var result ExampleProto
	err = rk.GetProto(ctx, key, &result)
	assert.NoError(t, err)
	assert.Equal(t, value, &result)

	// Ensure all expectations were met
	assert.NoError(t, mc.mock.ExpectationsWereMet())
}

func TestPing(t *testing.T) {
	mc, rk, err := NewMockRedisClient("json")
	assert.NoError(t, err)

	ctx := context.Background()

	// Mock PING
	mc.mock.ExpectPing().SetVal("PONG")

	err = rk.Ping(ctx)
	assert.NoError(t, err)

	assert.NoError(t, mc.mock.ExpectationsWereMet())
}

func TestDelete(t *testing.T) {
	mc, rk, err := NewMockRedisClient("json")
	assert.NoError(t, err)

	ctx := context.Background()
	key := "delete_key"

	// Mock DEL
	mc.mock.ExpectDel(key).SetVal(1)

	err = rk.Delete(ctx, key)
	assert.NoError(t, err)

	assert.NoError(t, mc.mock.ExpectationsWereMet())
}

func TestPipelineExample(t *testing.T) {
	mc, rk, err := NewMockRedisClient("json")
	assert.NoError(t, err)

	ctx := context.Background()

	// Mock INCR and EXPIRE in pipeline
	mc.mock.ExpectIncr("counter").SetVal(1)
	mc.mock.ExpectExpire("counter", time.Hour).SetVal(true)
	mc.mock.ExpectTxPipelineExec().SetVal([]interface{}{int64(1), true})

	err = rk.PipelineExample(ctx)
	assert.NoError(t, err)

	assert.NoError(t, mc.mock.ExpectationsWereMet())
}

func TestBulkSetAndGet(t *testing.T) {
	mc, rk, err := NewMockRedisClient("json")
	assert.NoError(t, err)

	ctx := context.Background()
	items := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	expiration := 5 * time.Minute

	// Mock SET operations in pipeline
	mc.mock.ExpectSet("key1", `"value1"`, expiration).SetVal("OK")
	mc.mock.ExpectSet("key2", `"value2"`, expiration).SetVal("OK")
	mc.mock.ExpectTxPipelineExec().SetVal([]interface{}{"OK", "OK"})

	err = rk.BulkSet(ctx, items, expiration)
	assert.NoError(t, err)

	// Mock GET operations in pipeline
	mc.mock.ExpectGet("key1").SetVal(`"value1"`)
	mc.mock.ExpectGet("key2").SetVal(`"value2"`)
	mc.mock.ExpectTxPipelineExec().SetVal([]interface{}{"value1", "value2"})

	dest := make(map[string]interface{})
	err = rk.BulkGet(ctx, []string{"key1", "key2"}, dest)
	assert.NoError(t, err)
	assert.Equal(t, "value1", dest["key1"])
	assert.Equal(t, "value2", dest["key2"])

	assert.NoError(t, mc.mock.ExpectationsWereMet())
}
