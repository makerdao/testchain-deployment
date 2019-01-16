package service

import (
	"fmt"
	"net/http"

	"github.com/makerdao/testchain-deployment/pkg/config"
	"github.com/makerdao/testchain-deployment/pkg/deploy"
	"github.com/makerdao/testchain-deployment/pkg/gateway"
	"github.com/makerdao/testchain-deployment/pkg/github"
	shttp "github.com/makerdao/testchain-deployment/pkg/service/http"
	"github.com/makerdao/testchain-deployment/pkg/service/methods"
	"github.com/makerdao/testchain-deployment/pkg/storage"
	"github.com/sirupsen/logrus"
)

//Run of service
func Run(log *logrus.Entry, cfg *config.Config) error {
	// init components
	gatewayClient := gateway.NewClient(cfg.Gateway)
	gatewayRegistrator := gateway.NewRegistrator(cfg.Gateway, gatewayClient, cfg.Host, cfg.Port)
	githubClient := github.NewClient(cfg.Github)
	inMemStorage := storage.NewInMemory()
	deployComponent := deploy.New(cfg.Deploy, githubClient, inMemStorage)
	methodsComponent := methods.NewMethods(inMemStorage, deployComponent, gatewayClient)
	// first load source
	if err := deployComponent.UpdateSource(log); err != nil {
		return err
	}
	// run assync registration on gateway and unrigistration on end of work
	go gatewayRegistrator.Run(log)
	defer gatewayRegistrator.Unregister(log)
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
	// init and run http server
	mux := http.NewServeMux()
	mux.Handle("/rpc", handler)

	httpServ := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}
	if err := httpServ.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
