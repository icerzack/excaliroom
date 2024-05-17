package room

import (
	"github.com/Icerzack/excalidraw-ws-go/internal/models"
)

const (
	InMemoryStorageType = "in-memory"
)

type Storage interface {
	Set(key string, value *models.Room) error
	Get(key string) (*models.Room, error)
	Delete(key string) error
}
