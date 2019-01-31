package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
	// register methods in handler
	handler := shttp.NewHandler(log)
	if err := handler.AddMethod("GetInfo", methodsComponent.GetInfo); err != nil {
		return err
	}
	if err := handler.AddMethod("Run", methodsComponent.Run); err != nil {
		return err
	}
	if err := handler.AddMethod("UpdateSource", methodsComponent.Update); err != nil {
		return err
	}
	if err := handler.AddMethod("GetResult", methodsComponent.GetResult); err != nil {
		return err
	}
	// init and run http server
	mux := http.NewServeMux()
	mux.Handle("/rpc", handler)

	httpServ := &HTTPServer{
		Storage: inMemStorage,
		Server: http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: mux,
		},
	}

	// operator for async group work and correct shutdown
	operator := system.NewOperator(
		log,
		httpServ,
		gatewayRegistrator,
	)
	signals := system.NewSignals(operator.GetErrCh())
	operator.Run()

	return signals.Wait(log, operator)
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
