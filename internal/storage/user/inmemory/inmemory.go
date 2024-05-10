package inmemory

import (
	"errors"
	"sync"

	"go.uber.org/zap"

	"github.com/Icerzack/excalidraw-ws-go/internal/user"
)

var ErrUserNotFound = errors.New("user not found")

type Storage struct {
	data   map[string]*user.User
	logger *zap.Logger

	mtx *sync.Mutex
}

func NewStorage(logger *zap.Logger) *Storage {
	return &Storage{
		data:   make(map[string]*user.User),
		logger: logger,
		mtx:    &sync.Mutex{},
	}
}

func (s *Storage) Set(key string, value *user.User) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.data[key] = value
	s.logger.Info("user added to storage", zap.String("key", key))
	return nil
}

func (s *Storage) Get(key string) (*user.User, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	v, ok := s.data[key]
	if !ok {
		s.logger.Info("user not found in storage", zap.String("key", key))
		return nil, ErrUserNotFound
	}
	return v, nil
}

func (s *Storage) Delete(key string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.data, key)
	s.logger.Info("user deleted from storage", zap.String("key", key))
	return nil
}

func (s *Storage) GetWhere(predicate func(*user.User) bool) (*user.User, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	for _, v := range s.data {
		if predicate(v) {
			return v, nil
		}
	}
	return nil, nil
}
