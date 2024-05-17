package models

import "github.com/gorilla/websocket"

// User is a struct that represents a user.
type User struct {
	// ID is the unique identifier of the user.
	ID string

	// RoomID is the unique identifier of the room that the user belongs to.
	RoomID string

	// Conn is the connection of the user.
	Conn *websocket.Conn
}
