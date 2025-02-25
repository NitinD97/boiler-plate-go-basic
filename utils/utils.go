package utils

import (
	"boiler-plate-go/constants"
	"boiler-plate-go/log"
	"context"
	"go.uber.org/zap"
)

func GetCtxLogger(ctx context.Context) *zap.Logger {
	return log.GetLogger().With(zap.String("requestID", ctx.Value(constants.RequestID).(string)))
}
