package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Icerzack/excaliroom/cmd"
	"github.com/Icerzack/excaliroom/internal/rest"
	"github.com/Icerzack/excaliroom/internal/utils"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")

	appConfig, err := cmd.ParseConfig(configPath)
	if err != nil {
		return
	}

	var level zapcore.Level
	var writeToFiles bool

	switch appConfig.Logging.Level {
	case "INFO":
		level = zap.InfoLevel
	case "DEBUG":
		level = zap.DebugLevel
	}

	if appConfig.Logging.WriteToFile {
		writeToFiles = true
	}

	logger, err := utils.NewCustomLogger(level, writeToFiles)
	if err != nil {
		fmt.Println("unable to create logger", err)
		return
	}

	restApp := rest.NewRest(&rest.Config{
		Port:               appConfig.Apps.Rest.Port,
		JwtValidationURL:   appConfig.Apps.Rest.Validation.JWTValidationURL,
		JwtHeaderName:      appConfig.Apps.Rest.Validation.JWTHeaderName,
		BoardValidationURL: appConfig.Apps.Rest.Validation.BoardValidationURL,
		UsersStorageType:   appConfig.Storage.Users.Type,
		RoomsStorageType:   appConfig.Storage.Rooms.Type,
		CacheType:          appConfig.Cache.Type,
		CacheTTL:           appConfig.Cache.TTL,
		Logger:             logger,
	})

	appsManager := cmd.NewAppsManager(logger)

	appsManager.Register(cmd.RestApp, restApp)
	appsManager.RunAll()
	appsManager.WaitForShutdown()
}
