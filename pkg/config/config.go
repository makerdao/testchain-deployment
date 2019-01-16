package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/makerdao/testchain-deployment/pkg/deploy"
	"github.com/makerdao/testchain-deployment/pkg/gateway"
	"github.com/makerdao/testchain-deployment/pkg/github"
)

// Config is an application config
type Config struct {
	Host     string         `split_word:"true"`
	Port     int            `split_word:"true"`
	Deploy   deploy.Config  `split_word:"true"`
	Gateway  gateway.Config `split_word:"true"`
	Github   github.Config  `split_word:"true"`
	LogLevel string         `split_word:"true"`
}

// EnvPrefix is prefix for env var, like a TCD_SOME_VAR
const EnvPrefix = "TCD"

// New init config with default params
func New() *Config {
	// set default values
	cfg := &Config{
		Host:     "testchain-deployment",
		Port:     5001,
		Deploy:   deploy.GetDefaultConfig(),
		Gateway:  gateway.GetDefaultConfig(),
		Github:   github.GetDefaultConfig(),
		LogLevel: "debug",
	}

	return cfg
}

// LoadFromEnv load configuration parameters from environment
func (c *Config) LoadFromEnv() error {
	return envconfig.Process(EnvPrefix, c)
}

// Validate cfg and all inclusion
// Return first error
func (c *Config) Validate() error {
	if err := c.Github.Validate(); err != nil {
		return err
	}

	return nil
}
