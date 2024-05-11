package main

import (
	"os"

	"go.uber.org/zap"

	"github.com/Icerzack/excalidraw-ws-go/cmd"
	"github.com/Icerzack/excalidraw-ws-go/internal/rest"
)

func main() {
	logger, _ := zap.NewDevelopment()
	//nolint:errcheck
	defer logger.Sync()

	configPath := os.Getenv("CONFIG_PATH")

	appConfig, err := cmd.ParseConfig(configPath, logger)
	if err != nil {
		return
	}

	switch appConfig.Apps.LogLevel {
	case "DEBUG":
		logger, _ = zap.NewDevelopment()
	case "INFO":
		logger, _ = zap.NewProduction()
	default:
		logger = zap.NewNop()
	}

	restApp := rest.NewRest(&rest.Config{
		Port:               appConfig.Apps.Rest.Port,
		JwtValidationURL:   appConfig.Apps.Rest.Validation.JWTValidationURL,
		JwtHeaderName:      appConfig.Apps.Rest.Validation.JWTHeaderName,
		BoardValidationURL: appConfig.Apps.Rest.Validation.BoardValidationURL,
		UsersStorageType:   appConfig.Storage.Users.Type,
		RoomsStorageType:   appConfig.Storage.Rooms.Type,
		Logger:             logger,
	})

	appsManager := cmd.NewAppsManager(logger)

	appsManager.Register(cmd.RestApp, restApp)
	appsManager.RunAll()
	appsManager.WaitForShutdown()
}
