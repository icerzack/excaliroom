package cmd

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

type AppsManager struct {
	apps map[string]App
	wg   *sync.WaitGroup

	logger *zap.Logger
}

func NewAppsManager(logger *zap.Logger) *AppsManager {
	return &AppsManager{
		apps:   make(map[string]App),
		wg:     &sync.WaitGroup{},
		logger: logger,
	}
}

func (am *AppsManager) Register(name string, app App) {
	am.apps[name] = app
}

func (am *AppsManager) Run(name string) {
	app, ok := am.apps[name]
	if !ok {
		return
	}
	am.logger.Info("App started", zap.String("name", name))
	app.Start()
}

func (am *AppsManager) Stop(name string) {
	app, ok := am.apps[name]
	if !ok {
		return
	}
	app.Stop()
	am.logger.Info("App stopped", zap.String("name", name))
}

func (am *AppsManager) RunAll() {
	for name, app := range am.apps {
		am.wg.Add(1)
		name := name
		go func(app App) {
			am.logger.Info("App started", zap.String("name", name))
			app.Start()
			am.wg.Done()
		}(app)
	}
}

func (am *AppsManager) StopAll() {
	for name, app := range am.apps {
		am.logger.Info("App stopped", zap.String("name", name))
		app.Stop()
	}
}

func (am *AppsManager) WaitForShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	am.StopAll()
	am.wg.Wait()
}
