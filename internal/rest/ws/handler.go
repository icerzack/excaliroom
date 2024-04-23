package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/Icerzack/excalidraw-ws-go/internal/room"
	rStorage "github.com/Icerzack/excalidraw-ws-go/internal/storage/room"
	uStorage "github.com/Icerzack/excalidraw-ws-go/internal/storage/user"
	"github.com/Icerzack/excalidraw-ws-go/internal/user"
)

var (
	ErrInvalidMessage = errors.New("invalid message")
	ErrValidatingJWT  = errors.New("failed to validate jwt")
)

const (
	EventConnect             = "connect"
	EventUserConnected       = "userConnected"
	EventUserDisconnected    = "userDisconnected"
	EventUserFailedToConnect = "userFailedToConnect"
	EventNewData             = "newData"
)

type WebSocketHandler struct {
	// upgrader is used to upgrade the HTTP connection to a WebSocket connection
	upgrader *websocket.Upgrader

	// jwtHeaderName is the name of the header that will be used to pass the JWT token
	jwtHeaderName string

	// jwtValidationURL is the URL that will be used to validate the JWT token
	jwtValidationURL string

	// userStorage is used to store the clients
	userStorage uStorage.Storage

	// roomStorage is used to store the rooms
	roomStorage rStorage.Storage

	logger *zap.Logger
}

func NewWebSocketHandler(
	clientsStorage uStorage.Storage,
	roomStorage rStorage.Storage,
	jwtHeaderName string,
	jwtValidationURL string,
	logger *zap.Logger,
) *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
		userStorage:      clientsStorage,
		roomStorage:      roomStorage,
		jwtHeaderName:    jwtHeaderName,
		jwtValidationURL: jwtValidationURL,
		logger:           logger,
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
		ws.logger.Info("Received MessageConnectRequest")
		ws.registerUser(conn, v)
	case MessageNewDataRequest:
		ws.logger.Info("Received MessageNewDataRequest")
		ws.sendDataToRoom(conn, v)
	}
}

func (ws *WebSocketHandler) sendDataToRoom(conn *websocket.Conn, request MessageNewDataRequest) {
	// Get the UserID from the JWT token
	userID, err := ws.validateJWT(request.Jwt)
	if err != nil {
		ws.logger.Debug("Failed to validate JWT", zap.Error(err))
		return
	}

	// Check if the user is registered
	if _, err := ws.userStorage.Get(userID); err != nil {
		return
	}

	// Check if user belongs to the room
	u, _ := ws.userStorage.Get(userID)
	if u == nil || u.RoomID != request.BoardID {
		ws.sendUserFailedToConnect(conn, "User not in this room")
		return
	}

	// Get the room
	currentRoom, _ := ws.roomStorage.Get(request.BoardID)
	if currentRoom == nil {
		return
	}

	// Update the current data
	currentRoom.SetElements(request.Data.Elements)
	currentRoom.SetAppState(request.Data.AppState)

	newData := Data{
		Elements: currentRoom.GetElements(),
		AppState: currentRoom.GetAppState(),
	}

	// Send the new data to all the users in the room
	for _, userID := range currentRoom.Users {
		u, _ := ws.userStorage.Get(userID.ID)
		if u == nil {
			continue
		}

		err := u.Conn.WriteJSON(MessageNewDataResponse{
			Message: Message{
				Event: EventNewData,
			},
			BoardID: currentRoom.BoardID,
			Data:    newData,
		})
		if err != nil {
			ws.unregisterUser(u.Conn)
		}
		ws.logger.Info("Data sent to user", zap.String("userID", u.ID))
	}
}

func (ws *WebSocketHandler) unregisterUser(conn *websocket.Conn) {
	// Get the user
	u, _ := ws.userStorage.GetWhere(func(u *user.User) bool {
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

	// Check if the room is empty
	if len(currentRoom.Users) == 0 {
		_ = ws.roomStorage.Delete(currentRoom.BoardID)
	}

	// Send the user disconnected message
	ws.sendUserDisconnected(MessageUserDisconnectedResponse{
		Message: Message{
			Event: EventUserDisconnected,
		},
		BoardID: currentRoom.BoardID,
		UserID:  u.ID,
	})

	// Remove the user from the storage
	_ = ws.userStorage.Delete(u.ID)
	ws.logger.Info("User unregistered", zap.String("userID", u.ID))
}

func (ws *WebSocketHandler) registerUser(conn *websocket.Conn, request MessageConnectRequest) {
	// Get the UserID from the JWT token
	userID, err := ws.validateJWT(request.Jwt)
	if err != nil {
		ws.logger.Debug("Failed to validate JWT", zap.Error(err))
		return
	}

	// Check if the user is already connected
	if v, _ := ws.userStorage.Get(userID); v != nil {
		return
	}

	// Create a room if it doesn't exist
	if v, _ := ws.roomStorage.Get(request.BoardID); v == nil {
		newRoom := room.NewRoom(request.BoardID)
		_ = ws.roomStorage.Set(request.BoardID, newRoom)
	}

	// Store the user
	newUser := &user.User{
		ID:     userID,
		RoomID: request.BoardID,
		Conn:   conn,
	}
	err = ws.userStorage.Set(newUser.ID, newUser)
	if err != nil {
		return
	}

	// Add the user to the room
	currentRoom, _ := ws.roomStorage.Get(request.BoardID)
	currentRoom.AddUser(newUser)

	// Send the user connected message
	ws.sendUserConnected(MessageUserConnectedResponse{
		Message: Message{
			Event: EventUserConnected,
		},
		BoardID: request.BoardID,
		UserID:  newUser.ID,
	})

	ws.logger.Info("User registered", zap.String("userID", newUser.ID))
}

func (ws *WebSocketHandler) sendUserConnected(request MessageUserConnectedResponse) {
	// Get the room
	currentRoom, _ := ws.roomStorage.Get(request.BoardID)
	if currentRoom == nil {
		return
	}

	// Send the message to all the users in the room
	for _, currentUser := range currentRoom.Users {
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

func (ws *WebSocketHandler) sendUserFailedToConnect(conn *websocket.Conn, reason string) {
	failedUser, err := ws.userStorage.GetWhere(func(u *user.User) bool {
		return u.Conn == conn
	})
	if err != nil {
		return
	}
	err = failedUser.Conn.WriteJSON(MessageUserFailedToConnectResponse{
		Message: Message{
			Event: EventUserFailedToConnect,
		},
		UserID: failedUser.ID,
		Reason: reason,
	})
	if err != nil {
		return
	}
	failedUser.Conn.Close()
}

func (ws *WebSocketHandler) sendUserDisconnected(request MessageUserDisconnectedResponse) {
	// Get the room
	currentRoom, _ := ws.roomStorage.Get(request.BoardID)
	if currentRoom == nil {
		return
	}

	// Send the message to all the users in the room
	for _, currentUser := range currentRoom.Users {
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
	req, _ := http.NewRequestWithContext(context.Background(), "GET", ws.jwtValidationURL, nil)
	req.Header.Set(ws.jwtHeaderName, jwt)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ws.logger.Error("Failed to send validation request", zap.Error(err))
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
	}
	return nil, ErrInvalidMessage
}
