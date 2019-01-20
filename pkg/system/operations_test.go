package system

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
)

type mock struct {
}

func (m *mock) Run(log *logrus.Entry) error {
	return nil
}

func (m *mock) Shutdown(context.Context, *logrus.Entry) error {
	return nil
}

func TestStubHandling(t *testing.T) {
	operator := NewOperator(logrus.WithField("component", "test"), &mock{})
	err := operator.Reload()
	if err != ErrNotImplemented {
		t.Error("Expected error", ErrNotImplemented, "got", err)
	}
	err = operator.Maintenance()
	if err != ErrNotImplemented {
		t.Error("Expected error", ErrNotImplemented, "got", err)
	}
	errs := operator.Shutdown()
	if len(errs) > 0 {
		t.Error("Expected success, got errors", errs)
	}
}
