package github

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Config for github API
type Config struct {
	APIToken              string
	RepoName              string
	TagName               string
	ClientTimeoutInSecond int
	WorkDir               string
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
		case "apiToken":
			c.APIToken = paramArr[1]
		case "repoName":
			c.RepoName = paramArr[1]
		case "tagName":
			c.TagName = paramArr[1]
		case "clientTimoutInSecond":
			v, err := strconv.Atoi(paramArr[1])
			if err != nil {
				return err
			}
			c.ClientTimeoutInSecond = v
		case "workDir":
			c.WorkDir = paramArr[1]
		default:
			return fmt.Errorf("unknown param '%s' for part of Githab env", paramArr[0])
		}
	}

	return nil
}

// Validate cfg after load
func (c *Config) Validate() error {
	if c.APIToken == "" {
		return errors.New("param github.APIToken can't be empty, use env var for set it")
	}

	return nil
}

// GetDefaultConfig return default config for github pkg
func GetDefaultConfig() Config {
	return Config{
		RepoName:              "testchain-dss-deployment-scripts",
		TagName:               "qa-deploy",
		ClientTimeoutInSecond: 10,
		WorkDir:               "/downloaded",
	}
}
