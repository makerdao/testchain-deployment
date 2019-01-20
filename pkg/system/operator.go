package system

import (
	"context"

	"github.com/sirupsen/logrus"
)

// Operator defines reload, maintenance and shutdown interface
type Operator interface {
	Reload() error
	Maintenance() error
	Shutdown() []error
}

// RunnerShutdowner defines Shutdown interface
type RunnerShutdowner interface {
	Run(log *logrus.Entry) error
	Shutdown(context.Context, *logrus.Entry) error
}
