package main

import (
	"go.uber.org/zap"
	"os"

	"github.com/Icerzack/excalidraw-ws-go/cmd"
	"github.com/Icerzack/excalidraw-ws-go/internal/rest"
)

func main() {
	logger, _ := zap.NewDevelopment()
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
		logger, _ = zap.NewDevelopment()
	}

	restApp := rest.NewRest(&rest.Config{
		Port:             appConfig.Apps.Rest.Port,
		JwtValidationURL: appConfig.Apps.Rest.JWT.ValidationURL,
		JwtHeaderName:    appConfig.Apps.Rest.JWT.HeaderName,
		UsersStorageType: appConfig.Storage.Users.Type,
		RoomsStorageType: appConfig.Storage.Rooms.Type,
		Logger:           logger,
	})

	appsManager := cmd.NewAppsManager(logger)

	appsManager.Register(cmd.RestApp, restApp)
	appsManager.RunAll()
	appsManager.WaitForShutdown()
}
