package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/makerdao/testchain-deployment/pkg/config"
	"github.com/makerdao/testchain-deployment/pkg/deploy"
	"github.com/makerdao/testchain-deployment/pkg/gateway"
	"github.com/makerdao/testchain-deployment/pkg/github"
	shttp "github.com/makerdao/testchain-deployment/pkg/service/http"
	"github.com/makerdao/testchain-deployment/pkg/service/methods"
	"github.com/makerdao/testchain-deployment/pkg/service/nats"
	"github.com/makerdao/testchain-deployment/pkg/storage"
	"github.com/makerdao/testchain-deployment/pkg/system"
	gonats "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
)

//Run of service
func Run(log *logrus.Entry, cfg *config.Config) error {
	// init components
	natsConn, err := gonats.Connect(
		cfg.NATS.Servers,
		gonats.MaxReconnects(cfg.NATS.MaxReconnect),
		gonats.ReconnectWait(time.Duration(cfg.NATS.ReconnectWaitSec)*time.Second),
	)
	if err != nil {
		return err
	}

	gatewayClient := gateway.NewClient(cfg.Gateway, natsConn, cfg.NATS)
	gatewayRegistrator := gateway.NewRegistrator(cfg.Gateway, gatewayClient, cfg.Host, cfg.Port)
	githubClient := github.NewClient(cfg.Github, cfg.Deploy.DeploymentDirPath)
	inMemStorage := storage.NewInMemory()
	deployComponent := deploy.New(cfg.Deploy, githubClient, inMemStorage)
	methodsComponent := methods.NewMethods(inMemStorage, deployComponent, gatewayClient)

	if err := deployComponent.FirstUpdate(log); err != nil {
		return err
	}

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
	if err := n.AddSyncMethod("GetCommitList", methodsComponent.GetCommitList); err != nil {
		return nil, err
	}
	if err := n.AddAsyncMethod("Run", methodsComponent.Run); err != nil {
		return nil, err
	}
	if err := n.AddAsyncMethod("UpdateSource", methodsComponent.Update); err != nil {
		return nil, err
	}
	if err := n.AddAsyncMethod("Checkout", methodsComponent.Checkout); err != nil {
		return nil, err
	}
	return n, nil
}

func httpServConfigure(
	log *logrus.Entry,
	port int,
	methodsComponent *methods.Methods,
	storage storage.Storage,
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
	if err := handler.AddMethod("GetCommitList", methodsComponent.GetCommitList); err != nil {
		return nil, err
	}
	if err := handler.AddMethod("Checkout", methodsComponent.Checkout); err != nil {
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

type HTTPServer struct {
	Storage storage.Storage
	http.Server
}

func (s *HTTPServer) Run(log *logrus.Entry) error {
	return s.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context, log *logrus.Entry) error {
	log.Debug("Start graceful shutdown http server")
	defer log.Debug("Graceful shutdown http server: done")
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
