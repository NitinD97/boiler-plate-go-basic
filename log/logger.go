package log

import (
	"boiler-plate-go/config"
	"boiler-plate-go/constants"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func Init() *zap.Logger {
	environment := config.GetString("environment")

	var err error
	if environment != "development" {
		writer := zapcore.AddSync(&lumberjack.Logger{
			Compress:   true,
			Filename:   config.GetString("log.fileName"),
			MaxSize:    config.GetInt("log.maxSize"), // megabytes
			MaxBackups: config.GetInt("log.maxBackups"),
			MaxAge:     config.GetInt("log.maxAge"), // days
		})
		location, _ := time.LoadLocation(constants.TIMEZONE)
		cfg := zap.NewProductionEncoderConfig()
		cfg.EncodeTime = func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
			t = t.In(location)
			zapcore.ISO8601TimeEncoder(t, pae)
		}
		cfg.EncodeDuration = zapcore.MillisDurationEncoder
		core := zapcore.NewTee(zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg),
			writer,
			zap.DebugLevel,
		))
		logger = zap.New(core)
	} else {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, err = cfg.Build()
	}
	if err != nil {
		panic(fmt.Errorf("unable to initialize logger\n %w", err))
	}

	defer logger.Sync()

	logger = logger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(0))
	//cronLogger = NewCronLogger(logger)
	return logger
}

func GetLogger() *zap.Logger {
	return logger
}

//func GetCronLogger() *CronLogger {
//	return cronLogger
//}
