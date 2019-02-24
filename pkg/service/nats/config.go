package nats

import (
	"fmt"
	"strconv"
	"strings"
)

// Config of gateway client
type Config struct {
	ErrorTopic       string
	GroupName        string
	TopicPrefix      string
	Servers          string
	MaxReconnect     int
	ReconnectWaitSec int
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
			return fmt.Errorf("bad param in part of GateWay env '%s'", p)
		}
		switch paramArr[0] {
		case "errorTopic":
			c.GroupName = paramArr[1]
		case "groupName":
			c.GroupName = paramArr[1]
		case "topicPrefix":
			c.TopicPrefix = paramArr[1]
		case "servers":
			c.Servers = paramArr[1]
		case "maxReconnect":
			v, err := strconv.Atoi(paramArr[1])
			if err != nil {
				return err
			}
			c.MaxReconnect = v
		case "reconnectWaitSec":
			v, err := strconv.Atoi(paramArr[1])
			if err != nil {
				return err
			}
			c.ReconnectWaitSec = v
		default:
			return fmt.Errorf("unknown param '%s' for part of Nats env", paramArr[0])
		}
	}

	return nil
}

// GetDefaultConfig return default config for gateway pkg
func GetDefaultConfig() Config {
	return Config{
		ErrorTopic:       "error",
		GroupName:        "testchain-deployment",
		TopicPrefix:      "Prefix",
		Servers:          "nats://nats.local:4222",
		MaxReconnect:     3,
		ReconnectWaitSec: 1,
	}
}
