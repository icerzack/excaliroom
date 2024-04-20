package user

import "github.com/Icerzack/excalidraw-ws-go/internal/user"

type Storage interface {
	Set(key string, value *user.User) error
	Get(key string) (*user.User, error)
	Delete(key string) error
	GetWhere(predicate func(*user.User) bool) (*user.User, error)
}
