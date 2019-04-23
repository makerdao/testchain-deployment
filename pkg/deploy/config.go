package deploy

import (
	"fmt"
	"strings"
)

// Config of deploy module
type Config struct {
	DeploymentDirPath string
	DeploymentSubPath string
	ResultSubPath     string
	RunUpdateOnStart  string
}

// Decode for envconfig
func (c *Config) Decode(data string) error {
	if data == "" {
		return nil
	}
	params := strings.Split(data, ";")
	for _, p := range params {
		paramArr := strings.Split(p, "=")
		if len(paramArr) != 2 {
			return fmt.Errorf("bad param in part of Deploy env '%s'", p)
		}
		switch paramArr[0] {
		case "deploymentDirPath":
			c.DeploymentDirPath = paramArr[1]
		case "deploymentSubPath":
			c.DeploymentSubPath = paramArr[1]
		case "resultSubPath":
			c.ResultSubPath = paramArr[1]
		case "runUpdateOnStart":
			c.RunUpdateOnStart = paramArr[1]
		default:
			return fmt.Errorf("unknown param '%s' for part of Deploy env", paramArr[0])
		}
	}

	return nil
}

// GetDefaultConfig return default config for local env
func GetDefaultConfig() Config {
	return Config{
		DeploymentDirPath: "/deployment",
		DeploymentSubPath: "./",
		ResultSubPath:     "out/addresses.json",
		RunUpdateOnStart:  "ifNotExists",
	}
}
