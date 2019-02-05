package github

import (
	"fmt"
	"strings"
)

// Config for github API
type Config struct {
	RepoOwner string
	RepoName  string
	TagName   string
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
			return fmt.Errorf("bad param in part of Githab env '%s'", p)
		}
		switch paramArr[0] {
		case "repoOwner":
			c.RepoOwner = paramArr[1]
		case "repoName":
			c.RepoName = paramArr[1]
		case "tagName":
			c.TagName = paramArr[1]
		default:
			return fmt.Errorf("unknown param '%s' for part of Githab env", paramArr[0])
		}
	}

	return nil
}

// Validate cfg after load
func (c *Config) Validate() error {
	return nil
}

// GetDefaultConfig return default config for github pkg
func GetDefaultConfig() Config {
	return Config{
		RepoOwner: "makerdao",
		RepoName:  "testchain-dss-deployment-scripts",
		TagName:   "qa-deploy",
	}
}
