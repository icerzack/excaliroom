package rest

import (
	"go.uber.org/zap"
)

type Config struct {
	// Port is the port where the server will listen
	Port int

	// JwtValidationURL is the URL which returns user id based on the JWT
	JwtValidationURL string

	Logger *zap.Logger
}
