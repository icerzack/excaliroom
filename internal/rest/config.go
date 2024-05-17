package rest

import (
	"go.uber.org/zap"
)

type Config struct {
	// Port is the port where the server will listen
	Port int

	// JwtHeaderName is the name of the header where the JWT is stored
	JwtHeaderName string

	// JwtValidationURL is the URL which returns user id based on the JWT
	JwtValidationURL string

	// BoardValidationURL is the URL which returns the board based on the board id
	BoardValidationURL string

	// UsersStorageType is the type of the storage that will be used
	UsersStorageType string

	// RoomsStorageType is the type of the storage that will be used
	RoomsStorageType string

	// CacheType is the type of the cache that will be used
	CacheType string

	// CacheTTL is the time to live of the cache
	CacheTTL int64

	Logger *zap.Logger
}
