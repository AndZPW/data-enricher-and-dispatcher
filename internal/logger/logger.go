package logger

import "go.uber.org/zap"

func MustInitLogger(env string) *zap.Logger {
	switch env {
	case "DEV":
		return zap.Must(zap.NewDevelopment())
	case "PROD":
		return zap.Must(zap.NewProduction())
	default:
		panic("no such environment: " + env)
	}
}
