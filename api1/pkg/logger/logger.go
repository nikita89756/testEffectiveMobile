package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)





type Logger interface {
	Error(string, ...zapcore.Field)
	Info(string, ...zapcore.Field)
	Fatal(string, ...zapcore.Field)
	Debug(string, ...zapcore.Field)
	Warn(string, ...zapcore.Field)
}



func InitLogger(console bool, filePath string,level string) (Logger, error) {
	var cores []zapcore.Core
	var levell zapcore.Level
	switch level {
	case "debug":
		levell = zap.DebugLevel
	case "info":
		levell = zap.InfoLevel
	case "warn":
		levell = zap.WarnLevel
	case "error":
		levell = zap.ErrorLevel
	default:
		levell = zap.InfoLevel
	}
	if console {
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), levell)
		cores = append(cores, consoleCore)
	}

	if filePath != "" {
		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil,err
		}
		fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(file), levell)
		cores = append(cores, fileCore)
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}

