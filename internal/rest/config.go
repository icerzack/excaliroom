package rest

import (
	"go.uber.org/zap"
)

type Config struct {
	// Port is the port where the server will listen
	Port int

	// JwtValidationURL is the URL which returns user id based on the JWT
	JwtValidationURL string

	// JwtHeaderName is the name of the header where the JWT is stored
	JwtHeaderName string

	// UsersStorageType is the type of the storage that will be used
	UsersStorageType string

	// RoomsStorageType is the type of the storage that will be used
	RoomsStorageType string

	Logger *zap.Logger
}
