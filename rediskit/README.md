# RedisKit

RedisKit is a comprehensive Redis client for Go, built on top of `github.com/redis/go-redis/v9`. It offers connection pooling, retry logic with exponential backoff, caching with `github.com/go-redis/cache/v8`, support for multiple encoding types (JSON, Msgpack, Protobuf), scalability through Redis clustering, performance optimizations like pipelining, robust error handling, health checks, and utilities for ease of use.

## Features

- **Connection Pooling**: Configurable pool size, idle connections, and timeouts.
- **Retry Logic**: Exponential backoff retries to handle transient failures.
- **Caching**: Easy-to-use caching with support for JSON, Msgpack, and Protobuf encodings.
- **Encoding Support**: Choose between JSON, Msgpack, Protobuf, and extend to other encodings.
- **Scalability**: Supports both standalone Redis and Redis Cluster configurations.
- **Performance Optimizations**: Supports pipelining and bulk operations for batch processing.
- **Error Handling**: Custom error handling for common Redis errors.
- **Health Checks**: Simple `Ping` method to verify Redis connectivity.
- **Utilities**: Easy methods for `Set`, `Get`, `Delete`, `Subscribe`, `Publish`, and more.
- **TLS/SSL Support**: Secure connections with TLS configuration.
- **Testing**: Includes unit tests with mocks for reliable testing without a live Redis server.

## Installation

```bash
go get github.com/xper626/go-buildingblocks/rediskit
