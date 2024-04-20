package rest

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/Icerzack/excalidraw-ws-go/internal/rest/ws"
	inmemRoom "github.com/Icerzack/excalidraw-ws-go/internal/storage/room/inmemory"
	inmemUser "github.com/Icerzack/excalidraw-ws-go/internal/storage/user/inmemory"
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
	usersStorage := inmemUser.NewStorage(rest.config.Logger)
	roomsStorage := inmemRoom.NewStorage(rest.config.Logger)
	wsServer := ws.NewWebSocketHandler(
		usersStorage,
		roomsStorage,
		"X-Auth-Token", // This is the header that will be used to pass the JWT token
		rest.config.JwtValidationURL,
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
	// Implement the stop logic
	if err := rest.server.Shutdown(context.Background()); err != nil {
		rest.config.Logger.Error("server error", zap.Error(err))
	}
}
