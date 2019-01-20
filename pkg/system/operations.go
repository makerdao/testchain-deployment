package system

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
)

// ErrNotImplemented declares error for method that isn't implemented
var ErrNotImplemented = errors.New("This method is not implemented")

// ErrEmptyServerPointer declares error for nil pointer
var ErrEmptyServerPointer = errors.New("Server pointer should not be nil")

// Operations implements simplest Operator interface
type Operations struct {
	errCh chan error
	list  []RunnerShutdowner
	log   *logrus.Entry
}

// NewOperator creates operator
func NewOperator(log *logrus.Entry, sd ...RunnerShutdowner) *Operations {
	service := &Operations{
		log:   log.WithField("component", "operator"),
		list:  make([]RunnerShutdowner, 0),
		errCh: make(chan error),
	}
	service.list = append(service.list, sd...)

	return service
}

func (o *Operations) GetErrCh() <-chan error {
	return o.errCh
}

func (o *Operations) Run() {
	for _, rs := range o.list {
		go func(rs RunnerShutdowner, errCh chan<- error) {
			if err := rs.Run(o.log); err != nil {
				errCh <- err
			}
		}(rs, o.errCh)
	}
}

// Reload operation implementation
func (o Operations) Reload() error {
	return ErrNotImplemented
}

// Maintenance operation implementation
func (o Operations) Maintenance() error {
	return ErrNotImplemented
}

// Shutdown operation
func (o Operations) Shutdown() []error {
	var errs []error
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Minute)
	defer cancel()
	errCh := make(chan error)
	expectedCount := len(o.list)
	for _, fn := range o.list {
		go func(fn RunnerShutdowner, errCh chan<- error) {
			errCh <- fn.Shutdown(ctx, o.log)
		}(fn, errCh)
	}

	for err := range errCh {
		if err != nil {
			errs = append(errs, err)
		}
		expectedCount--
		if expectedCount == 0 {
			return errs
		}
	}
	return errs
}
