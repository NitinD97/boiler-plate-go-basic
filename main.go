package main

import (
	"boiler-plate-go/config"
	"boiler-plate-go/log"
	"go.uber.org/zap"
	"net/http"
)

type App struct {
	logger *zap.Logger
	server *http.Server
}

func main() {
	config.Init()
	logger := log.Init()

	app := App{
		logger: logger,
	}

	app.InitializeServer()
}
