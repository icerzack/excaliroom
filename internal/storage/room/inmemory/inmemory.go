package inmemory

import (
	"errors"
	"go.uber.org/zap"

	"github.com/Icerzack/excalidraw-ws-go/internal/room"
)

var (
	ErrRoomNotFound = errors.New("room not found")
)

type Storage struct {
	data   map[string]*room.Room
	logger *zap.Logger
}

func NewStorage(logger *zap.Logger) *Storage {
	return &Storage{
		data:   make(map[string]*room.Room),
		logger: logger,
	}
}

func (s *Storage) Set(key string, value *room.Room) error {
	s.data[key] = value
	s.logger.Info("room added to storage", zap.String("key", key))
	return nil
}

func (s *Storage) Get(key string) (*room.Room, error) {
	v, ok := s.data[key]
	if !ok {
		s.logger.Info("room not found in storage", zap.String("key", key))
		return nil, ErrRoomNotFound
	}
	return v, nil
}

func (s *Storage) Delete(key string) error {
	delete(s.data, key)
	s.logger.Info("room deleted from storage", zap.String("key", key))
	return nil
}
