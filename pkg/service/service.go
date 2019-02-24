package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/makerdao/testchain-deployment/pkg/service/nats"

	"github.com/makerdao/testchain-deployment/pkg/config"
	"github.com/makerdao/testchain-deployment/pkg/deploy"
	"github.com/makerdao/testchain-deployment/pkg/gateway"
	"github.com/makerdao/testchain-deployment/pkg/github"
	shttp "github.com/makerdao/testchain-deployment/pkg/service/http"
	"github.com/makerdao/testchain-deployment/pkg/service/methods"
	"github.com/makerdao/testchain-deployment/pkg/storage"
	"github.com/makerdao/testchain-deployment/pkg/system"
	"github.com/sirupsen/logrus"
)

//Run of service
func Run(log *logrus.Entry, cfg *config.Config) error {
	// init components
	gatewayClient := gateway.NewClient(cfg.Gateway)
	gatewayRegistrator := gateway.NewRegistrator(cfg.Gateway, gatewayClient, cfg.Host, cfg.Port)
	githubClient := github.NewClient(cfg.Github, cfg.Deploy.DeploymentDirPath)
	inMemStorage := storage.NewInMemory()
	deployComponent := deploy.New(cfg.Deploy, githubClient, inMemStorage)
	methodsComponent := methods.NewMethods(inMemStorage, deployComponent, gatewayClient)
	// first load source
	log.Info("First update src started, it takes a few minutes")
	if err := deployComponent.UpdateSource(log); err != nil {
		log.WithError(err).Error("Can't first update source")
		return err
	}
	log.Info("First update src finished")

	log.Infof("Used %s server", cfg.Server)
	var serv system.RunnerShutdowner
	switch cfg.Server {
	case "HTTP":
		var err error
		serv, err = httpServConfigure(log, cfg.Port, methodsComponent, inMemStorage)
		if err != nil {
			return err
		}
	case "NATS":
		var err error
		serv, err = natsServConfigure(log, cfg.NATS, methodsComponent)
		if err != nil {
			return err
		}
	default:
		return errors.New("server can be only HTTP or NATS")
	}

	// operator for async group work and correct shutdown
	operator := system.NewOperator(
		log,
		serv,
		gatewayRegistrator,
	)
	signals := system.NewSignals(operator.GetErrCh())
	operator.Run()

	return signals.Wait(log, operator)
}

func natsServConfigure(log *logrus.Entry, cfg nats.Config, methodsComponent *methods.Methods) (*nats.Server, error) {
	n := nats.New(log, &cfg)

	if err := n.AddSyncMethod("GetInfo", methodsComponent.GetInfo); err != nil {
		return nil, err
	}
	if err := n.AddSyncMethod("GetResult", methodsComponent.GetResult); err != nil {
		return nil, err
	}
	return n, nil
}

func httpServConfigure(
	log *logrus.Entry,
	port int,
	methodsComponent *methods.Methods,
	storage HTTPServerStorage,
) (*HTTPServer, error) {
	// register methods in handler
	handler := shttp.NewHandler(log)
	if err := handler.AddMethod("GetInfo", methodsComponent.GetInfo); err != nil {
		return nil, err
	}
	if err := handler.AddMethod("Run", methodsComponent.Run); err != nil {
		return nil, err
	}
	if err := handler.AddMethod("UpdateSource", methodsComponent.Update); err != nil {
		return nil, err
	}
	if err := handler.AddMethod("GetResult", methodsComponent.GetResult); err != nil {
		return nil, err
	}
	// init and run http server
	mux := http.NewServeMux()
	mux.Handle("/rpc", handler)

	return &HTTPServer{
		Storage: storage,
		Server: http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}, nil
}

type HTTPServerStorage interface {
	GetUpdate() bool
	GetRun() bool
}

type HTTPServer struct {
	Storage HTTPServerStorage
	http.Server
}

func (s *HTTPServer) Run(log *logrus.Entry) error {
	return s.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context, log *logrus.Entry) error {
	// stop http server for new request
	err := s.Server.Shutdown(ctx)
	//wait while all operations will be finished
	opCh := make(chan struct{})
	go func() {
		for {
			if !s.Storage.GetRun() {
				opCh <- struct{}{}
				return
			}
			time.Sleep(10 * time.Second)
		}
	}()
	select {
	case <-opCh:
		return err
	case <-ctx.Done():
		return fmt.Errorf("context cancalled, but operations not colmpleted, server err: %s", err.Error())
	}
}
