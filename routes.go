package main

import (
	"github.com/gin-gonic/gin"
)

func (app *App) registerRoutes(router *gin.Engine) {

	router.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

}
