package gateway

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// Registrator is a component for registration instance on gateway
type Registrator struct {
	client         *Client
	tickerDuration time.Duration
	host           string
	port           int
	registered     bool
}

// NewRegistrator init registrator
func NewRegistrator(cfg Config, client *Client, host string, port int) *Registrator {
	return &Registrator{
		client:         client,
		tickerDuration: time.Duration(cfg.RegisterPeriodInSec) * time.Second,
		host:           host,
		port:           port,
	}
}

// Run registrator
func (r *Registrator) Run(log *logrus.Entry) error {
	ticker := time.NewTicker(r.tickerDuration)
	//nolint:megacheck
	for {
		select {
		case <-ticker.C:
			err := r.client.Register(
				log.WithField("component", "gateway_client"),
				&ServiceData{
					Host: r.host,
					Port: r.port,
				},
			)
			if err != nil {
				log.WithError(err).Warn("Can't register instance on gateway")
				continue
			}
			r.registered = true
			return nil
		}
	}
}

//Shutdown unregister from gateway
func (r *Registrator) Shutdown(ctx context.Context, log *logrus.Entry) error {
	if !r.registered {
		log.Info("Deployment was not registered")
		return nil
	}
	err := r.client.Unregister(log, &ServiceData{
		Host: r.host,
		Port: r.port,
	})
	if err != nil {
		log.WithError(err).Error("Can't unregister deployment")
		return err
	}
	log.Info("Deployment unregistered")
	return nil
}
