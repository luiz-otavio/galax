package config

import (
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	logging struct {
		Debug bool
	}

	api struct {
		Key string
	}

	mysql struct {
		DSN string
	}

	redis struct {
		DSN      string
		Interval int64
		Key      string
	}

	server struct {
		Binding string
	}

	worker struct {
		Parallelism int
		Interval    int
		Iterations  int
	}
}

func Load(file string) (*Config, error) {
	var config Config

	_, err := toml.DecodeFile(file, &config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) GetParallelism() int {
	return c.worker.Parallelism
}

func (c *Config) GetRedis() string {
	return c.redis.DSN
}

func (c *Config) GetInterval() int {
	return c.worker.Interval
}

func (c *Config) GetIterations() int {
	return c.worker.Iterations
}

func (c *Config) GetBinding() string {
	return c.server.Binding
}

func (c *Config) GetMySQL() string {
	return c.mysql.DSN
}

func (c *Config) GetKey() string {
	return c.api.Key
}

func (c *Config) GetDebug() bool {
	return c.logging.Debug
}

func (c *Config) GetExpireInterval() time.Duration {
	return time.Duration(c.redis.Interval)
}

func (c *Config) GetAccountKey() string {
	return c.redis.Key
}
