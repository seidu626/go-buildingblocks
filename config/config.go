package config

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/redis/go-redis/v9"
	"github.com/seidu626/go-buildingblocks/database"
	"github.com/seidu626/go-buildingblocks/database/cassandra"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Environment string

const (
	DEVELOPMENT Environment = "DEVELOPMENT"
	PRODUCTION              = "PRODUCTION"
)

type Config struct {
	Application struct {
		Name           string              `mapstructure:"NAME"`
		Environment    Environment         `mapstructure:"ENVIRONMENT"`
		Port           int                 `mapstructure:"PORT"`
		AllowedOrigins []string            `mapstructure:"ALLOWED_ORIGINS"`
		TelcoPrefixes  map[string][]string `mapstructure:"TELCO_PREFIXES"`
		TIMWE          struct {
			Host              string        `mapstructure:"HOST"`
			BaseURL           string        `mapstructure:"BASE_URL"`
			APIKey            string        `mapstructure:"API_KEY"`
			PartnerID         string        `mapstructure:"PARTNER_ID"`
			Realm             string        `mapstructure:"REALM"`
			AuthenticationKey string        `mapstructure:"AUTHENTICATION_KEY"`
			Timeout           time.Duration `mapstructure:"TIMEOUT"`
			MaxConnections    int           `mapstructure:"MAX_CONNECTIONS"`
		} `mapstructure:"TIMWE_MA"`
		Log struct {
			Path string `mapstructure:"PATH"`
		}
		Key struct {
			Default string `mapstructure:"DEFAULT"`
			Rsa     struct {
				Public  string `mapstructure:"PUBLIC"`
				Private string `mapstructure:"PRIVATE"`
			}
		}
		Graceful struct {
			MaxSecond time.Duration `mapstructure:"MAX_SECOND"`
		} `mapstructure:"GRACEFUL"`
	} `mapstructure:"APPLICATION"`
	Auth struct {
		model        string
		ClientID     string
		ClientSecret string
		Audience     string
		JwtToken     struct {
			Type           string `mapstructure:"TYPE"`
			Expired        string `mapstructure:"EXPIRED"`
			Secret         string `mapstructure:"SECRET"`
			RefreshExpired string `mapstructure:"REFRESH_EXPIRED"`
		} `mapstructure:"JWT_TOKEN"`
	} `mapstructure:"AUTH"`
	DB struct {
		Postgresql struct {
			DBHost     string `mapstructure:"HOST"`
			DBPort     string `mapstructure:"PORT"`
			DBUser     string `mapstructure:"USER"`
			DBPassword string `mapstructure:"PASSWORD"`
			DBName     string `mapstructure:"DB_NAME"`
			SSLMode    string `mapstructure:"SSL_MODE"`
		} `mapstructure:"POSTGRESQL"`
		Cassandra struct { // New Cassandra configuration
			Hosts          []string      `mapstructure:"HOSTS"`
			Port           int           `mapstructure:"PORT"`
			Username       string        `mapstructure:"USERNAME"`
			Password       string        `mapstructure:"PASSWORD"`
			Keyspace       string        `mapstructure:"KEYSPACE"`
			Consistency    string        `mapstructure:"CONSISTENCY"`
			Timeout        time.Duration `mapstructure:"TIMEOUT"`
			ConnectTimeout time.Duration `mapstructure:"CONNECT_TIMEOUT"`
			PoolSize       int           `mapstructure:"POOL_SIZE"`
			RetryPolicy    string        `mapstructure:"RETRY_POLICY"`
			ProtoVersion   int           `mapstructure:"PROTO_VERSION"`
			SSL            bool          `mapstructure:"SSL"`
			SslOpts        struct {      // Nested struct for SSL options
				EnableHostVerification bool     `mapstructure:"ENABLE_HOST_VERIFICATION"`
				CaPath                 string   `mapstructure:"CA_PATH"`
				CertPath               string   `mapstructure:"CERT_PATH"`
				KeyPath                string   `mapstructure:"KEY_PATH"`
				Ciphers                []string `mapstructure:"CIPHERS"`
				ServerName             string   `mapstructure:"SERVER_NAME"`
			} `mapstructure:"SSL_OPTIONS"`
		} `mapstructure:"CASSANDRA"`
		CockroachDB struct {
			Regions map[string]struct {
				Hosts      []string `mapstructure:"HOSTS"`
				Port       int      `mapstructure:"PORT"`
				Username   string   `mapstructure:"USERNAME"`
				Password   string   `mapstructure:"PASSWORD"`
				Database   string   `mapstructure:"DATABASE"`
				SSL        bool     `mapstructure:"SSL"`
				SSLOptions struct { // Nested struct for SSL options
					EnableHostVerification bool     `mapstructure:"ENABLE_HOST_VERIFICATION"`
					CaPath                 string   `mapstructure:"CA_PATH"`
					CertPath               string   `mapstructure:"CERT_PATH"`
					KeyPath                string   `mapstructure:"KEY_PATH"`
					Ciphers                []string `mapstructure:"CIPHERS"`
					ServerName             string   `mapstructure:"SERVER_NAME"`
				} `mapstructure:"SSL_OPTIONS"`
				PoolSize        database.PoolSize `mapstructure:"POOL_SIZE"`
				ConnectTimeout  time.Duration     `mapstructure:"CONNECT_TIMEOUT"`
				MaxIdleConns    int               `mapstructure:"MAX_IDLE_CONNS"`
				MaxOpenConns    int               `mapstructure:"MAX_OPEN_CONNS"`
				ConnMaxLifetime time.Duration     `mapstructure:"CONN_MAX_LIFETIME"`
			} `mapstructure:"REGIONS"`
		} `mapstructure:"COCKROACHDB"`
	} `mapstructure:"DB"`
	Cache struct {
		Redis struct {
			Host              string   `mapstructure:"HOST"`
			Port              int      `mapstructure:"PORT"`
			DB                int      `mapstructure:"DB"`
			Pass              string   `mapstructure:"PASS"`
			PoolSize          int      `mapstructure:"POOL_SIZE"`
			MinIdleConns      int      `mapstructure:"MIN_IDLE_CONNS"`
			IdleTimeout       string   `mapstructure:"IDLE_TIMEOUT"`       // Use string for duration parsing
			DefaultExpiration string   `mapstructure:"DEFAULT_EXPIRATION"` // Use string for duration parsing
			MaxRetries        int      `mapstructure:"MAX_RETRIES"`
			MinRetryBackoff   string   `mapstructure:"MIN_RETRY_BACKOFF"`
			MaxRetryBackoff   string   `mapstructure:"MAX_RETRY_BACKOFF"`
			Encoding          string   `mapstructure:"ENCODING"`
			IsCluster         bool     `mapstructure:"IS_CLUSTER"`
			Addrs             []string `mapstructure:"ADDRESSES"` // Redis Cluster addresses
			TLSEnabled        bool     `mapstructure:"TLS_ENABLED"`
			TLSCaPath         string   `mapstructure:"TLS_CA_PATH"`
			TLSCertPath       string   `mapstructure:"TLS_CERT_PATH"`
			TLSKeyPath        string   `mapstructure:"TLS_KEY_PATH"`
			TLSServerName     string   `mapstructure:"TLS_SERVER_NAME"`
			TLSEnableHostname bool     `mapstructure:"TLS_ENABLE_HOSTNAME"`
		} `mapstructure:"REDIS"`
	} `mapstructure:"CACHE"`
	Logging struct {
		Level  string `mapstructure:"LEVEL"`
		Format string `mapstructure:"FORMAT"`
	} `mapstructure:"LOGGING"`
	// DynamicConfigs holds configurations registered by external services
	DynamicConfigs map[string]interface{}
}

// Global configuration pointer and mutex
var (
	cfg      *Config
	cfgMutex sync.RWMutex
)

// Registry to hold configuration structs registered by external packages
var (
	registry = make(map[string]interface{})
	regMutex = sync.Mutex{}
)

// RegisterConfig allows external packages to register their configuration struct with a key.
func RegisterConfig(key string, configStruct interface{}) {
	regMutex.Lock()
	defer regMutex.Unlock()
	registry[key] = configStruct
}

// GetDynamicConfig returns the registered configuration for the given key.
func GetDynamicConfig(key string) (interface{}, bool) {
	cfgMutex.RLock()
	defer cfgMutex.RUnlock()
	cfgStruct, exists := cfg.DynamicConfigs[key]
	return cfgStruct, exists
}

// GetDynamicConfigTyped returns the registered configuration for the given key with type assertion.
func GetDynamicConfigTyped[T any](key string) (*T, bool) {
	cfgMutex.RLock()
	defer cfgMutex.RUnlock()
	cfgStruct, exists := cfg.DynamicConfigs[key]
	if !exists {
		return nil, false
	}
	typedCfg, ok := cfgStruct.(*T)
	return typedCfg, ok
}

// InitConfig initializes the configuration by loading config files and setting up watchers.
func InitConfig(logger *zap.Logger, path string, files []string) *Config {
	v := viper.New()
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetConfigType("yaml") // or "json", depending on your config files
	cfg = &Config{}
	cfg.DynamicConfigs = make(map[string]interface{})

	// Load configuration files
	for i, file := range files {
		v.SetConfigFile(fmt.Sprintf("%s/%s", path, file))
		if i == 0 {
			// Read the first config file
			if err := v.ReadInConfig(); err != nil {
				logger.Warn("Config file load error (continuing)", zap.Error(err))
			}
		} else {
			// Merge subsequent config files
			if err := v.MergeInConfig(); err != nil {
				logger.Warn("Config file merge error (continuing)", zap.Error(err))
			}
		}
	}

	// Unmarshal to config struct
	if err := v.Unmarshal(cfg); err != nil {
		logger.Fatal("Error unmarshalling config", zap.Error(err))
	}

	// Unmarshal registered configurations
	regMutex.Lock()
	for key, cfgStruct := range registry {
		if err := v.UnmarshalKey(key, cfgStruct); err != nil {
			logger.Error("Error unmarshalling registered config", zap.String("key", key), zap.Error(err))
		} else {
			cfg.DynamicConfigs[key] = cfgStruct
		}
	}
	regMutex.Unlock()

	// Watch for changes
	v.OnConfigChange(func(e fsnotify.Event) {
		logger.Info("Config file changed", zap.String("file", e.Name))
		newCfg := &Config{}
		if err := v.Unmarshal(newCfg); err != nil {
			logger.Error("Error reloading config", zap.Error(err))
			return
		}

		// Update the global config in a thread-safe manner
		cfgMutex.Lock()
		*cfg = *newCfg
		cfgMutex.Unlock()

		// Reload registered configurations
		regMutex.Lock()
		for key, cfgStruct := range registry {
			if err := v.UnmarshalKey(key, cfgStruct); err != nil {
				logger.Error("Error reloading registered config", zap.String("key", key), zap.Error(err))
			} else {
				cfgMutex.Lock()
				cfg.DynamicConfigs[key] = cfgStruct
				cfgMutex.Unlock()
				logger.Info("Dynamic config reloaded", zap.String("key", key))
			}
		}
		regMutex.Unlock()
	})
	v.WatchConfig()

	return cfg
}

// GetDBConnectionString constructs the PostgreSQL connection string from the config.
func GetDBConnectionString() string {
	cfgMutex.RLock()
	defer cfgMutex.RUnlock()
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Postgresql.DBHost,
		cfg.DB.Postgresql.DBPort,
		cfg.DB.Postgresql.DBUser,
		cfg.DB.Postgresql.DBPassword,
		cfg.DB.Postgresql.DBName,
		cfg.DB.Postgresql.SSLMode,
	)
	return connStr
}

// GetRedisOptions constructs the Redis options from the config.
func GetRedisOptions() *redis.Options {
	cfgMutex.RLock()
	defer cfgMutex.RUnlock()
	return &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Cache.Redis.Host, cfg.Cache.Redis.Port),
		Password: cfg.Cache.Redis.Pass,
		DB:       cfg.Cache.Redis.DB,
	}
}

// GetCockroachDBConfig retrieves the CockroachDB configuration for a specific region.
func GetCockroachDBConfig(region string) (*database.Config, error) {
	cfgMutex.RLock()
	defer cfgMutex.RUnlock()

	regionConfig, exists := cfg.DB.CockroachDB.Regions[region]
	if !exists {
		return &database.Config{}, fmt.Errorf("no CockroachDB configuration found for region: %s", region)
	}

	// Construct the DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		regionConfig.Username,
		regionConfig.Password,
		strings.Join(regionConfig.Hosts, ","),
		regionConfig.Port,
		regionConfig.Database,
	)

	if regionConfig.SSL {
		dsn += "?sslmode=verify-full"
		// Append SSL parameters if needed
	} else {
		dsn += "?sslmode=disable"
	}

	return &database.Config{
		Hosts:           regionConfig.Hosts,
		Port:            regionConfig.Port,
		Username:        regionConfig.Username,
		Password:        regionConfig.Password,
		Database:        regionConfig.Database,
		SSL:             regionConfig.SSL,
		PoolSize:        regionConfig.PoolSize,
		ConnectTimeout:  regionConfig.ConnectTimeout,
		MaxIdleConns:    regionConfig.MaxIdleConns,
		MaxOpenConns:    regionConfig.MaxOpenConns,
		ConnMaxLifetime: regionConfig.ConnMaxLifetime,
	}, nil
}

// GetCassandraConfig constructs the Cassandra configuration from the config package.
func GetCassandraConfig() *cassandra.Config {
	cfgMutex.RLock()
	defer cfgMutex.RUnlock()

	// Map string consistency to gocql.Consistency
	var consistency gocql.Consistency
	switch strings.ToUpper(cfg.DB.Cassandra.Consistency) {
	case "ANY":
		consistency = gocql.Any
	case "ONE":
		consistency = gocql.One
	case "TWO":
		consistency = gocql.Two
	case "THREE":
		consistency = gocql.Three
	case "QUORUM":
		consistency = gocql.Quorum
	case "ALL":
		consistency = gocql.All
	default:
		consistency = gocql.Quorum // Default consistency
	}

	// Map string retry policy to gocql.RetryPolicy
	var retryPolicy gocql.RetryPolicy
	switch strings.ToUpper(cfg.DB.Cassandra.RetryPolicy) {
	case "SIMPLE":
		retryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}
	case "EXPONENTIAL_BACKOFF":
		retryPolicy = &gocql.ExponentialBackoffRetryPolicy{
			NumRetries: 3,
			Min:        100 * time.Millisecond,
			Max:        1 * time.Second,
		}
	//case "DOWNSCALING":
	//	// For DowngradingConsistencyRetryPolicy, we need to specify the underlying retry policy
	//	// and the consistency level to downgrade to. For example, downgrade to ONE consistency.
	//	retryPolicy = gocql.DowngradingConsistencyRetryPolicy{
	//		Policy:  &gocql.SimpleRetryPolicy{NumRetries: 3},
	//		RetryTo: gocql.One,
	//	}
	case "DEFAULT":
		// Assuming "DEFAULT" maps to the built-in DefaultRetryPolicy
		retryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}
	default:
		// Default to SimpleRetryPolicy if unknown
		retryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}
	}

	return &cassandra.Config{
		Hosts:          cfg.DB.Cassandra.Hosts,
		Port:           cfg.DB.Cassandra.Port,
		Username:       cfg.DB.Cassandra.Username,
		Password:       cfg.DB.Cassandra.Password,
		Keyspace:       cfg.DB.Cassandra.Keyspace,
		Consistency:    consistency,
		Timeout:        cfg.DB.Cassandra.Timeout,
		ConnectTimeout: cfg.DB.Cassandra.ConnectTimeout,
		PoolSize:       cfg.DB.Cassandra.PoolSize,
		RetryPolicy:    retryPolicy,
		ProtoVersion:   cfg.DB.Cassandra.ProtoVersion,
		SSL:            cfg.DB.Cassandra.SSL,
		SslOpts: &gocql.SslOptions{
			EnableHostVerification: cfg.DB.Cassandra.SslOpts.EnableHostVerification,
			CaPath:                 cfg.DB.Cassandra.SslOpts.CaPath,
			CertPath:               cfg.DB.Cassandra.SslOpts.CertPath,
			KeyPath:                cfg.DB.Cassandra.SslOpts.KeyPath,
			//Ciphers:                cfg.DB.Cassandra.SslOpts.Ciphers,
			//ServerName:             cfg.DB.Cassandra.SslOpts.ServerName,
		},
	}
}
