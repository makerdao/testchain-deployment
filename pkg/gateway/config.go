package gateway

import (
	"fmt"
	"strconv"
	"strings"
)

// Config of gateway client
type Config struct {
	Host                  string
	Port                  int
	ClientTimeoutInSecond int
	RegisterPeriodInSec   int
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
		case "host":
			c.Host = paramArr[1]
		case "port":
			v, err := strconv.Atoi(paramArr[1])
			if err != nil {
				return err
			}
			c.Port = v
		case "clientTimeoutInSecond":
			v, err := strconv.Atoi(paramArr[1])
			if err != nil {
				return err
			}
			c.ClientTimeoutInSecond = v
		case "registerPeriodInSec":
			v, err := strconv.Atoi(paramArr[1])
			if err != nil {
				return err
			}
			c.RegisterPeriodInSec = v
		default:
			return fmt.Errorf("unknown param '%s' for part of GateWay env", paramArr[0])
		}
	}

	return nil
}

// GetDefaultConfig return default config for gateway pkg
func GetDefaultConfig() Config {
	return Config{
		Host:                  "testchain-backendgateway",
		Port:                  4000,
		ClientTimeoutInSecond: 5,
		RegisterPeriodInSec:   10,
	}
}
