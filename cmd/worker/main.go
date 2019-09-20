package main

import (
	"log"

	"github.com/makerdao/testchain-deployment/pkg/worker"

	"github.com/makerdao/testchain-deployment/pkg/config"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.New()
	if err := cfg.LoadFromEnv(); err != nil {
		log.Fatalln(err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalln(err)
	}

	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatalln(err)
	}
	logrus.SetLevel(level)
	logger := logrus.WithField("app", config.EnvPrefix)
	logger.Info("Config loaded")
	logger.Debugf("Config: %+v", cfg)

	logger.Infof("Start service with host: %s, port: %d", cfg.Host, cfg.Port)
	if err := worker.Execute(logger, cfg); err != nil {
		log.Fatalln(err)
	}
	logger.Info("Application finished")
}
