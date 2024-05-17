package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Icerzack/excalidraw-ws-go/internal/cache"
	"github.com/Icerzack/excalidraw-ws-go/internal/models"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	rStorage "github.com/Icerzack/excalidraw-ws-go/internal/storage/room"
	uStorage "github.com/Icerzack/excalidraw-ws-go/internal/storage/user"
)

var (
	ErrInvalidMessage = errors.New("invalid message")
	ErrValidatingJWT  = errors.New("failed to validate jwt")
)

const (
	EventConnect          = "connect"
	EventUserConnected    = "userConnected"
	EventUserDisconnected = "userDisconnected"
	EventSetLeader        = "setLeader"
	EventNewData          = "newData"
)

type WebSocketHandler struct {
	// upgrader is used to upgrade the HTTP connection to a WebSocket connection
	upgrader *websocket.Upgrader

	// jwtHeaderName is the name of the header that will be used to pass the JWT token
	jwtHeaderName string

	// jwtValidationURL is the URL that will be used to validate the JWT token
	jwtValidationURL string

	// boardValidationURL is the URL that will be used to validate the board access
	boardValidationURL string

	// userStorage is used to store the clients
	userStorage uStorage.Storage

	// roomStorage is used to store the rooms
	roomStorage rStorage.Storage

	// cache is used to store the validation results
	cache cache.Cache

	logger *zap.Logger
}

func NewWebSocketHandler(
	clientsStorage uStorage.Storage,
	roomStorage rStorage.Storage,
	cache cache.Cache,
	jwtHeaderName string,
	jwtValidationURL string,
	boardValidationURL string,
	logger *zap.Logger,
) *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
		userStorage:        clientsStorage,
		roomStorage:        roomStorage,
		jwtHeaderName:      jwtHeaderName,
		jwtValidationURL:   jwtValidationURL,
		boardValidationURL: boardValidationURL,
		cache:              cache,
		logger:             logger,
	}
}

func (ws *WebSocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}
	defer conn.Close()
	ws.logger.Info("Connection upgraded successfully")
	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil || mt == websocket.CloseMessage {
			ws.unregisterUser(conn)
			ws.logger.Info("Connection closed")
			break
		}

		// Handle the incoming message
		go ws.messageHandler(conn, msg)
	}
}

func (ws *WebSocketHandler) messageHandler(conn *websocket.Conn, msg []byte) {
	message, err := messageDefiner(msg)
	if err != nil {
		ws.logger.Debug("Failed to define message", zap.Error(err))
		return
	}

	switch v := message.(type) {
	case MessageConnectRequest:
		ws.registerUser(conn, v)
	case MessageNewDataRequest:
		ws.sendDataToRoom(v)
	case MessageSetLeaderRequest:
		ws.setLeader(v)
	}
}

//nolint:cyclop
func (ws *WebSocketHandler) setLeader(request MessageSetLeaderRequest) {
	var userID string

	// Check if the user is in cache
	if v, err := ws.cache.Get(request.Jwt); err != nil && v == nil {
		// Get the UserID from the JWT token
		userID, err = ws.validateJWT(request.Jwt)
		if err != nil {
			ws.logger.Debug("Failed to validate JWT", zap.Error(err))
			return
		}

		// Check if the user has access to the board
		if !ws.validateBoardAccess(request.BoardID, request.Jwt) {
			ws.logger.Debug("User doesn't have access to the board", zap.String("userID", userID), zap.String("boardID", request.BoardID))
			return
		}

		// Store the validation result
		_ = ws.cache.Set(request.Jwt, userID)
	} else {
		userID = v.(string)
	}

	// Check if the user is registered
	u, err := ws.userStorage.Get(userID)
	if err != nil {
		return
	}

	// Check if user belongs to the room
	if u == nil || u.RoomID != request.BoardID {
		return
	}

	// Get the room
	currentRoom, _ := ws.roomStorage.Get(request.BoardID)
	if currentRoom == nil {
		return
	}

	// Set the leader
	switch currentRoom.LeaderID {
	case "0":
		currentRoom.SetLeader(userID)
		ws.logger.Debug("User set as the leader", zap.String("userID", userID), zap.String("boardID", request.BoardID))
	case userID:
		currentRoom.SetLeader("0")
		ws.logger.Debug("Leader removed", zap.String("userID", userID), zap.String("boardID", request.BoardID))
	default:
		return
	}

	// Send the message to all the users in the room
	for _, currentUser := range currentRoom.GetUsers() {
		u, _ := ws.userStorage.Get(currentUser.ID)
		if u == nil {
			continue
		}

		err := u.Conn.WriteJSON(MessageSetLeaderResponse{
			Message: Message{
				Event: EventSetLeader,
			},
			BoardID: request.BoardID,
			UserID:  currentRoom.LeaderID,
		})
		if err != nil {
			// Remove the user from the storage
			_ = ws.userStorage.Delete(currentUser.ID)
		}
	}
}

//nolint:cyclop
func (ws *WebSocketHandler) sendDataToRoom(request MessageNewDataRequest) {
	var userID string

	// Check if the user is in cache
	if v, err := ws.cache.Get(request.Jwt); err != nil && v == nil {
		// Get the UserID from the JWT token
		userID, err = ws.validateJWT(request.Jwt)
		if err != nil {
			ws.logger.Debug("Failed to validate JWT", zap.Error(err))
			return
		}

		// Check if the user has access to the board
		if !ws.validateBoardAccess(request.BoardID, request.Jwt) {
			ws.logger.Debug("User doesn't have access to the board", zap.String("userID", userID), zap.String("boardID", request.BoardID))
			return
		}

		// Store the validation result
		_ = ws.cache.Set(request.Jwt, userID)
	} else {
		userID = v.(string)
	}

	// Check if the user is registered
	if _, err := ws.userStorage.Get(userID); err != nil {
		return
	}

	// Check if user belongs to the room
	u, _ := ws.userStorage.Get(userID)
	if u == nil || u.RoomID != request.BoardID {
		return
	}

	// Get the room
	currentRoom, _ := ws.roomStorage.Get(request.BoardID)
	if currentRoom == nil {
		return
	}

	currentRoom.RoomMutex.Lock()
	defer currentRoom.RoomMutex.Unlock()

	// Check if the user is the leader
	if currentRoom.LeaderID != userID {
		return
	}

	// Update the current data
	currentRoom.SetElements(request.Data.Elements)
	currentRoom.SetAppState(request.Data.AppState)

	ws.logger.Debug("Data updated", zap.String("userID", userID), zap.String("boardID", currentRoom.BoardID))

	// Send the new data to all the users in the room
	for _, v := range currentRoom.GetUsers() {
		u, _ := ws.userStorage.Get(v.ID)
		if u == nil {
			continue
		}

		err := u.Conn.WriteJSON(MessageNewDataResponse{
			Message: Message{
				Event: EventNewData,
			},
			BoardID: currentRoom.BoardID,
			Data: Data{
				Elements: currentRoom.GetElements(),
				AppState: currentRoom.GetAppState(),
			},
		})
		if err != nil {
			ws.unregisterUser(u.Conn)
		}
	}
}

func (ws *WebSocketHandler) unregisterUser(conn *websocket.Conn) {
	// Get the user
	u, _ := ws.userStorage.GetWhere(func(u *models.User) bool {
		return u.Conn == conn
	})
	if u == nil {
		return
	}

	// Get the room
	currentRoom, _ := ws.roomStorage.Get(u.RoomID)
	if currentRoom == nil {
		return
	}

	// Remove the user from the room
	currentRoom.RemoveUser(u.ID)

	// Check if the user was the leader
	if currentRoom.LeaderID == u.ID {
		currentRoom.SetLeader("0")
	}

	// Remove the user from the storage
	_ = ws.userStorage.Delete(u.ID)
	ws.logger.Info("User unregistered", zap.String("userID", u.ID))

	// Check if the room is empty
	if len(currentRoom.GetUsers()) == 0 {
		_ = ws.roomStorage.Delete(currentRoom.BoardID)
		return
	}

	// Get the users ids
	userIDs := make([]string, 0)
	for _, u := range currentRoom.GetUsers() {
		userIDs = append(userIDs, u.ID)
	}

	// Send the user disconnected message
	ws.sendUserDisconnected(MessageUserDisconnectedResponse{
		Message: Message{
			Event: EventUserDisconnected,
		},
		BoardID:  currentRoom.BoardID,
		UserIDs:  userIDs,
		LeaderID: currentRoom.LeaderID,
	})
}

func (ws *WebSocketHandler) registerUser(conn *websocket.Conn, request MessageConnectRequest) {
	var userID string

	// Check if the user is in cache
	if v, err := ws.cache.Get(request.Jwt); err != nil && v == nil {
		// Get the UserID from the JWT token
		userID, err = ws.validateJWT(request.Jwt)
		if err != nil {
			ws.logger.Debug("Failed to validate JWT", zap.Error(err))
			return
		}

		// Check if the user has access to the board
		if !ws.validateBoardAccess(request.BoardID, request.Jwt) {
			ws.logger.Debug("User doesn't have access to the board", zap.String("userID", userID), zap.String("boardID", request.BoardID))
			return
		}

		// Store the validation result
		_ = ws.cache.Set(request.Jwt, userID)
	} else {
		userID = v.(string)
	}

	// Check if the user is already connected
	if v, _ := ws.userStorage.Get(userID); v != nil {
		return
	}

	// Create a room if it doesn't exist
	var currentRoom *models.Room
	if currentRoom, _ = ws.roomStorage.Get(request.BoardID); currentRoom == nil {
		currentRoom = models.NewRoom(request.BoardID)
		_ = ws.roomStorage.Set(request.BoardID, currentRoom)
	}

	// Store the user
	newUser := &models.User{
		ID:     userID,
		RoomID: request.BoardID,
		Conn:   conn,
	}
	err := ws.userStorage.Set(newUser.ID, newUser)
	if err != nil {
		return
	}

	// Add the user to the room
	currentRoom.AddUser(newUser)

	// Get the users ids
	userIDs := make([]string, 0)
	for _, u := range currentRoom.GetUsers() {
		userIDs = append(userIDs, u.ID)
	}

	// Send the user connected message
	ws.sendUserConnected(MessageUserConnectedResponse{
		Message: Message{
			Event: EventUserConnected,
		},
		BoardID:  request.BoardID,
		UserIDs:  userIDs,
		LeaderID: currentRoom.LeaderID,
	})

	ws.logger.Info("User registered", zap.String("userID", newUser.ID), zap.String("roomID", newUser.RoomID))
}

func (ws *WebSocketHandler) sendUserConnected(request MessageUserConnectedResponse) {
	// Get the room
	currentRoom, _ := ws.roomStorage.Get(request.BoardID)
	if currentRoom == nil {
		return
	}

	// Send the message to all the users in the room
	for _, currentUser := range currentRoom.GetUsers() {
		u, _ := ws.userStorage.Get(currentUser.ID)
		if u == nil {
			continue
		}

		err := u.Conn.WriteJSON(request)
		if err != nil {
			// Remove the user from the storage
			_ = ws.userStorage.Delete(currentUser.ID)
		}
	}
}

func (ws *WebSocketHandler) sendUserDisconnected(request MessageUserDisconnectedResponse) {
	// Get the room
	currentRoom, _ := ws.roomStorage.Get(request.BoardID)
	if currentRoom == nil {
		return
	}

	// Send the message to all the users in the room
	for _, currentUser := range currentRoom.GetUsers() {
		u, _ := ws.userStorage.Get(currentUser.ID)
		if u == nil {
			continue
		}

		err := u.Conn.WriteJSON(request)
		if err != nil {
			// Remove the user from the storage
			_ = ws.userStorage.Delete(currentUser.ID)
		}
	}
}

func (ws *WebSocketHandler) validateJWT(jwt string) (string, error) {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, ws.jwtValidationURL, nil)
	req.Header.Set(ws.jwtHeaderName, jwt)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ws.logger.Error("Failed to send jwt validation request", zap.Error(err))
		return "", fmt.Errorf("failed to send validation request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return "", fmt.Errorf("unauthorized: %w", ErrValidatingJWT)
	case http.StatusForbidden:
		return "", fmt.Errorf("forbidden: %w", ErrValidatingJWT)
	case http.StatusInternalServerError:
		return "", fmt.Errorf("internal server error: %w", ErrValidatingJWT)
	}

	// Decode the JWTValidationResponse
	var jwtResponse JWTValidationResponse
	err = json.NewDecoder(resp.Body).Decode(&jwtResponse)
	if err != nil {
		ws.logger.Error("Failed to decode JWT response", zap.Error(err))
		return "", fmt.Errorf("failed to decode JWT response: %w", err)
	}

	return strconv.Itoa(jwtResponse.ID), nil
}

func (ws *WebSocketHandler) validateBoardAccess(boardID, jwt string) bool {
	fullURL, err := url.JoinPath(ws.boardValidationURL, boardID)
	if err != nil {
		ws.logger.Error("Failed to join URL", zap.Error(err))
		return false
	}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, fullURL, nil)
	req.Header.Set(ws.jwtHeaderName, jwt)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ws.logger.Error("Failed to send board validation request", zap.Error(err))
		return false
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true
	default:
		return false
	}
}

func messageDefiner(msg []byte) (interface{}, error) {
	var message Message
	if err := json.Unmarshal(msg, &message); err != nil {
		return nil, ErrInvalidMessage
	}
	switch message.Event {
	case EventConnect:
		var connectRequest MessageConnectRequest
		if err := json.Unmarshal(msg, &connectRequest); err == nil {
			return connectRequest, nil
		} else {
			return nil, fmt.Errorf("error Unmarshaling MessageConnectRequest: %w", err)
		}
	case EventNewData:
		var newData MessageNewDataRequest
		if err := json.Unmarshal(msg, &newData); err == nil {
			return newData, nil
		} else {
			return nil, fmt.Errorf("error Unmarshaling MessageNewDataRequest: %w", err)
		}
	case EventSetLeader:
		var setLeaderRequest MessageSetLeaderRequest
		if err := json.Unmarshal(msg, &setLeaderRequest); err == nil {
			return setLeaderRequest, nil
		} else {
			return nil, fmt.Errorf("error Unmarshaling MessageSetLeaderRequest: %w", err)
		}
	}
	return nil, ErrInvalidMessage
}
