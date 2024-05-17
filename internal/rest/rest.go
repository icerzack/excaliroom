package rest

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Icerzack/excaliroom/internal/cache"
	"github.com/Icerzack/excaliroom/internal/cache/inmemory"
	"github.com/Icerzack/excaliroom/internal/rest/ws"
	"github.com/Icerzack/excaliroom/internal/storage/room"
	inmemRoom "github.com/Icerzack/excaliroom/internal/storage/room/inmemory"
	"github.com/Icerzack/excaliroom/internal/storage/user"
	inmemUser "github.com/Icerzack/excaliroom/internal/storage/user/inmemory"
)

type Rest struct {
	config *Config

	server *http.Server
}

func NewRest(config *Config) *Rest {
	return &Rest{
		config: config,
	}
}

func (rest *Rest) Start() {
	router := chi.NewRouter()

	// Define the /ping endpoint
	router.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("pong"))
		if err != nil {
			return
		}
	})

	// Define the /ws endpoint
	usersStorage, roomsStorage := rest.defineStorage()
	selectedCache := rest.defineCache()

	wsServer := ws.NewWebSocketHandler(
		usersStorage,
		roomsStorage,
		selectedCache,
		rest.config.CacheTTL,
		rest.config.JwtHeaderName,
		rest.config.JwtValidationURL,
		rest.config.BoardValidationURL,
		rest.config.Logger,
	)
	router.HandleFunc("/ws", wsServer.Handle)

	rest.server = &http.Server{
		Addr:              ":" + strconv.Itoa(rest.config.Port),
		Handler:           router,
		ReadHeaderTimeout: 0,
	}
	if err := rest.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		rest.config.Logger.Error("server error", zap.Error(err))
		return
	}
}

func (rest *Rest) Stop() {
	if err := rest.server.Shutdown(context.Background()); err != nil {
		rest.config.Logger.Error("server error", zap.Error(err))
	}
}

func (rest *Rest) defineStorage() (user.Storage, room.Storage) {
	var usersStorage user.Storage
	var roomsStorage room.Storage

	switch rest.config.UsersStorageType {
	case user.InMemoryStorageType:
		rest.config.Logger.Info("Using in-memory storage for users")
		usersStorage = inmemUser.NewStorage(rest.config.Logger)
	default:
		rest.config.Logger.Info("Using in-memory storage for users")
		usersStorage = inmemUser.NewStorage(rest.config.Logger)
	}
	switch rest.config.RoomsStorageType {
	case room.InMemoryStorageType:
		rest.config.Logger.Info("Using in-memory storage for rooms")
		roomsStorage = inmemRoom.NewStorage(rest.config.Logger)
	default:
		rest.config.Logger.Info("Using in-memory storage for rooms")
		roomsStorage = inmemRoom.NewStorage(rest.config.Logger)
	}

	return usersStorage, roomsStorage
}

func (rest *Rest) defineCache() cache.Cache {
	var c cache.Cache

	switch rest.config.CacheType {
	case room.InMemoryStorageType:
		rest.config.Logger.Info("Using in-memory cache")
		c = inmemory.NewCache(rest.config.Logger)
	default:
		rest.config.Logger.Info("Using in-memory cache")
		c = inmemory.NewCache(rest.config.Logger)
	}

	return c
}
