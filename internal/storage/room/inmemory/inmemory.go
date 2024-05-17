package inmemory

import (
	"errors"
	"github.com/Icerzack/excalidraw-ws-go/internal/models"
	"sync"

	"go.uber.org/zap"
)

var ErrRoomNotFound = errors.New("room not found")

type Storage struct {
	data   map[string]*models.Room
	logger *zap.Logger

	mtx *sync.Mutex
}

func NewStorage(logger *zap.Logger) *Storage {
	return &Storage{
		data:   make(map[string]*models.Room),
		logger: logger,
		mtx:    &sync.Mutex{},
	}
}

func (s *Storage) Set(key string, value *models.Room) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.data[key] = value
	s.logger.Info("room added to storage", zap.String("key", key))
	return nil
}

func (s *Storage) Get(key string) (*models.Room, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	v, ok := s.data[key]
	if !ok {
		s.logger.Info("room not found in storage", zap.String("key", key))
		return nil, ErrRoomNotFound
	}
	return v, nil
}

func (s *Storage) Delete(key string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.data, key)
	s.logger.Info("room deleted from storage", zap.String("key", key))
	return nil
}
