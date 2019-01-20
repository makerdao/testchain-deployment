package methods

import (
	"github.com/makerdao/testchain-deployment/pkg/deploy"
	"github.com/makerdao/testchain-deployment/pkg/gateway"
	"github.com/sirupsen/logrus"
)

//StorageInterface for methods
type StorageInterface interface {
	GetStepList(log *logrus.Entry) ([]deploy.StepModel, error)
	GetTagHash(log *logrus.Entry) (hash string, err error)
	HasData() bool
	GetRun() bool
	GetUpdate() bool
}

//Methods is main methods struct as container for DI
type Methods struct {
	storage         StorageInterface
	deployComponent *deploy.Component
	gatewayClient   *gateway.Client
}

//NewMethods init methods
func NewMethods(
	storage StorageInterface,
	deployComponent *deploy.Component,
	gatewayClient *gateway.Client,
) *Methods {
	return &Methods{storage: storage, deployComponent: deployComponent, gatewayClient: gatewayClient}
}
