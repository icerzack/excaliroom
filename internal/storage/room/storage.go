package room

import (
	"github.com/Icerzack/excalidraw-ws-go/internal/room"
)

const (
	InMemoryStorageType = "in-memory"
)

type Storage interface {
	Set(key string, value *room.Room) error
	Get(key string) (*room.Room, error)
	Delete(key string) error
}
