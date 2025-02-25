package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trinkerr/referral-service/config"
	"github.com/trinkerr/referral-service/constants"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

func requestIdInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := c.Request.Header.Get("x-request-id")
		if requestId == "" {
			requestId = uuid.NewString()
		}

		r := c.Request.Context()
		updatedContext := context.WithValue(r, constants.RequestID, requestId)
		c.Request = c.Request.WithContext(updatedContext)
		c.Next()
	}
}

// GinLogger receives the default log of the gin framework
func ginLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		logger.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("requestId", c.Request.Context().Value(constants.RequestID).(string)),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

// GinRecovery removes the possible panic of the project
func ginRecovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, true)
				if brokenPipe {
					logger.Sugar().Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("requestId", c.Request.Context().Value(constants.RequestID).(string)))
					logger.Sugar().Error(string(httpRequest))
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error))
					c.Abort()
					return
				}
				logger.Sugar().Error(err)
				logger.Sugar().Info(string(debug.Stack()))
				logger.Sugar().Error("[Recovery from panic]", string(httpRequest))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"errorMessage": "Something went wrong, please try again later",
				})
			}
		}()
		c.Next()
	}
}

func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := config.GetSlice("cors.allowedOrigin")
	env := config.GetString("environment")

	return func(c *gin.Context) {
		isAllowed := false
		origin := c.Request.Header.Get("Origin")
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == origin {
				isAllowed = true
				break
			}
		}
		if isAllowed || env == "staging" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Referer, Authorization")
		c.Header("Access-Control-Allow-Methods", "POST,GET,OPTIONS,DELETE,PUT,PATCH")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "600")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
