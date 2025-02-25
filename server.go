package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/trinkerr/referral-service/config"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var router *gin.Engine

func (app *App) InitializeServer() {
	environment := config.GetString("environment")
	port := config.GetString("server.port")
	readTimeout := config.GetInt("server.readTimeout")
	writeTimeout := config.GetInt("server.writeTimeout")

	if environment != "development" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router = gin.New()
	router.Use(gzip.Gzip(gzip.BestSpeed))
	router.UseH2C = true
	router.Use(requestIdInterceptor())
	router.Use(ginLogger(app.logger), ginRecovery(app.logger))
	router.Use(corsMiddleware())

	app.registerRoutes(router)
	router.HandleMethodNotAllowed = true

	h2s := &http2.Server{}
	app.server = &http.Server{
		Addr:           port,
		Handler:        h2c.NewHandler(router, h2s),
		ReadTimeout:    time.Duration(readTimeout) * time.Second,
		WriteTimeout:   time.Duration(writeTimeout) * time.Second,
		MaxHeaderBytes: 1 << 10,
	}
	app.cleanup()

	app.logger.Info(fmt.Sprintf("Starting server on port %s", port))
	if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		app.logger.Fatal("failed to start server", zap.Error(err))
	}
}

func (app *App) cleanup() {
	closureChan := make(chan os.Signal, 1)
	signal.Notify(closureChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-closureChan
		app.logger.Info("Received shutdown signal, shutting down...")
		err := app.server.Shutdown(context.Background())
		if err != nil {
			app.logger.Error("failed to shutdown server", zap.Error(err))
		} else {
			app.logger.Info("Server gracefully stopped")
		}
		app.logger.Sync()
		os.Exit(1)
	}()
}
