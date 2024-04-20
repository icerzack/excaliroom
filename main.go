package main

import (
	"go.uber.org/zap"

	"github.com/Icerzack/excalidraw-ws-go/cmd"
	"github.com/Icerzack/excalidraw-ws-go/internal/rest"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	restApp := rest.NewRest(&rest.Config{
		Port:             8080,
		JwtValidationURL: "https://api.camelot.icerzack.space/api/v1/users/me",
		Logger:           logger,
	})

	appsManager := cmd.NewAppsManager(logger)

	appsManager.Register(cmd.RestApp, restApp)
	appsManager.RunAll()
	appsManager.WaitForShutdown()
}
