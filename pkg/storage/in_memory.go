package storage

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/makerdao/testchain-deployment/pkg/deploy"
	"github.com/sirupsen/logrus"
)

//InMemory implementation of storage, use mutex for data consistent
type InMemory struct {
	hasHash     bool
	hasManifest bool
	run         bool
	update      bool
	hash        string
	manifest    deploy.Manifest
	mu          sync.Mutex
	updatedAt   time.Time
}

//NewInMemory init storaga
func NewInMemory() *InMemory {
	return &InMemory{}
}

func (s *InMemory) UpsertManifest(log *logrus.Entry, manifest deploy.Manifest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.manifest = manifest
	s.hasManifest = true
	return nil
}

func (s *InMemory) GetManifest(log *logrus.Entry) (*deploy.Manifest, error) {
	if !s.HasData() {
		return nil, errors.New("has not loaded data")
	}
	return &s.manifest, nil
}

func (s *InMemory) GetScenario(log *logrus.Entry, scenarioNr int) (*deploy.Scenario, error) {
	hasScenario, err := s.HasScenario(log, scenarioNr)
	if err != nil {
		return nil, errors.New("has not loaded data")
	}
	if !hasScenario {
		return nil, fmt.Errorf("scenario nr. %d not available", scenarioNr)
	}
	return &s.manifest.Scenarios[scenarioNr-1], nil
}

func (s *InMemory) GetStepList(log *logrus.Entry) ([]deploy.StepModel, error) {
	if !s.HasData() {
		return nil, errors.New("has not loaded data")
	}
	stepList, err := deploy.NewStepListFromManifest(&s.manifest)
	if err != nil {
		return nil, err
	}
	return stepList, nil
}

func (s *InMemory) HasScenario(log *logrus.Entry, scenarioNr int) (bool, error) {
	if !s.HasData() {
		return false, errors.New("has not loaded data")
	}
	return (1 <= scenarioNr && scenarioNr <= len(s.manifest.Scenarios)), nil
}

func (s *InMemory) SetTagHash(log *logrus.Entry, hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hash = hash
	s.hasHash = true
	return nil
}

func (s *InMemory) GetTagHash(log *logrus.Entry) (hash string, err error) {
	return s.hash, nil
}

func (s *InMemory) HasData() bool {
	return s.hasHash && s.hasManifest
}

func (s *InMemory) SetRun(run bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.run && run {
		return errors.New("deployment already run")
	}
	s.run = run
	return nil
}

func (s *InMemory) GetRun() bool {
	return s.run
}

func (s *InMemory) SetUpdate(update bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.update && update {
		return errors.New("update source already run")
	}
	s.update = update
	return nil
}

func (s *InMemory) GetUpdate() bool {
	return s.update
}

func (s *InMemory) SetUpdatedAtNow() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updatedAt = time.Now()
	return nil
}

func (s *InMemory) GetUpdatedAt() (*time.Time, error) {
	return &s.updatedAt, nil
}
