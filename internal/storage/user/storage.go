package user

import (
	"github.com/Icerzack/excalidraw-ws-go/internal/models"
)

const (
	InMemoryStorageType = "in-memory"
)

type Storage interface {
	Set(key string, value *models.User) error
	Get(key string) (*models.User, error)
	Delete(key string) error
	GetWhere(predicate func(*models.User) bool) (*models.User, error)
}
