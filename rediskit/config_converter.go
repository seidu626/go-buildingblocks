package rediskit

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/seidu626/go-buildingblocks/config"
)

// ConvertToRedisKitConfig converts a general config.Config to rediskit.Config
func ConvertToRedisKitConfig(appConfig *config.Config) (Config, error) {
	redisConf := appConfig.Cache.Redis

	// Validate Redis configuration
	if !redisConf.IsCluster && redisConf.Host == "" {
		return Config{}, fmt.Errorf("redis configuration error: Host is required for non-cluster mode")
	}
	if redisConf.IsCluster && len(redisConf.Addrs) == 0 {
		return Config{}, fmt.Errorf("redis configuration error: At least one address is required for cluster mode")
	}

	// Construct the Addr field
	var addr string
	if !redisConf.IsCluster {
		addr = fmt.Sprintf("%s:%d", redisConf.Host, redisConf.Port)
	} else {
		// In cluster mode, Addr is not used. Addrs slice will be utilized instead.
		addr = ""
	}

	// Parse duration strings into time.Duration
	idleTimeout, err := time.ParseDuration(redisConf.IdleTimeout)
	if err != nil {
		return Config{}, fmt.Errorf("invalid IdleTimeout duration: %v", err)
	}

	minRetryBackoff, err := time.ParseDuration(redisConf.MinRetryBackoff)
	if err != nil {
		return Config{}, fmt.Errorf("invalid MinRetryBackoff duration: %v", err)
	}

	maxRetryBackoff, err := time.ParseDuration(redisConf.MaxRetryBackoff)
	if err != nil {
		return Config{}, fmt.Errorf("invalid MaxRetryBackoff duration: %v", err)
	}

	// Parse other duration fields if present in appConfig.Cache
	defaultExpiration, err := time.ParseDuration(redisConf.DefaultExpiration)
	if defaultExpiration <= 0 {
		defaultExpiration = 10 * time.Minute // Default value
	}

	// Configure TLS if enabled
	var tlsConfig *tls.Config
	if redisConf.TLSEnabled {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: !redisConf.TLSEnableHostname,
			ServerName:         redisConf.TLSServerName,
		}

		// Load client certificate if provided
		if redisConf.TLSCertPath != "" && redisConf.TLSKeyPath != "" {
			cert, err := tls.LoadX509KeyPair(redisConf.TLSCertPath, redisConf.TLSKeyPath)
			if err != nil {
				return Config{}, fmt.Errorf("failed to load TLS key pair: %v", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		// Load CA certificate if provided
		if redisConf.TLSCaPath != "" {
			caCert, err := ioutil.ReadFile(redisConf.TLSCaPath)
			if err != nil {
				return Config{}, fmt.Errorf("failed to read TLS CA certificate: %v", err)
			}
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				return Config{}, fmt.Errorf("failed to append CA certificate")
			}
			tlsConfig.RootCAs = caCertPool
		}
	}

	// Set default values for DialTimeout, ReadTimeout, WriteTimeout if not specified
	dialTimeout := 5 * time.Second
	readTimeout := 3 * time.Second
	writeTimeout := 3 * time.Second

	// Initialize rediskit.Config with mapped fields
	redisKitConf := Config{
		Addr:         addr,
		Password:     redisConf.Pass,
		DB:           redisConf.DB,
		PoolSize:     redisConf.PoolSize,
		MinIdleConns: redisConf.MinIdleConns,
		IdleTimeout:  idleTimeout,

		DialTimeout:  dialTimeout,  // Can be modified to parse from config if available
		ReadTimeout:  readTimeout,  // Can be modified to parse from config if available
		WriteTimeout: writeTimeout, // Can be modified to parse from config if available
		RetryCount:   redisConf.MaxRetries,
		RetryDelay:   minRetryBackoff,

		MaxRetries:      redisConf.MaxRetries,
		MinRetryBackoff: minRetryBackoff,
		MaxRetryBackoff: maxRetryBackoff,

		DefaultExpiration: defaultExpiration,

		Encoding: redisConf.Encoding,

		IsCluster: redisConf.IsCluster,
		Addrs:     redisConf.Addrs,

		TLSConfig: tlsConfig,
	}

	return redisKitConf, nil
}
