package storage

import (
	"errors"
	"sync"

	"github.com/makerdao/testchain-deployment/pkg/deploy"
	"github.com/sirupsen/logrus"
)

//InMemory implementation of storage, use mutex for data consistent
type InMemory struct {
	hasHash  bool
	hasList  bool
	run      bool
	update   bool
	hash     string
	modelMap map[int]deploy.Model
	mu       sync.Mutex
}

//NewInMemory init storaga
func NewInMemory() *InMemory {
	return &InMemory{
		modelMap: make(map[int]deploy.Model),
	}
}

func (s *InMemory) UpsertStepList(log *logrus.Entry, modelList []deploy.Model) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modelMap = make(map[int]deploy.Model)
	for _, model := range modelList {
		if _, ok := s.modelMap[model.ID]; ok {
			return errors.New("some model has equal id")
		}
		s.modelMap[model.ID] = model
	}
	s.hasList = true
	return nil
}

func (s *InMemory) GetStepList(log *logrus.Entry) ([]deploy.Model, error) {
	if !s.HasData() {
		return nil, errors.New("has not loaded data")
	}
	res := make([]deploy.Model, 0)
	for _, model := range s.modelMap {
		res = append(res, model)
	}
	return res, nil
}

func (s *InMemory) HasStepList(log *logrus.Entry, id int) (bool, error) {
	return len(s.modelMap) > 0, nil
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
	return s.hasHash && s.hasList
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
