package logging

import (
	"os"

	"github.com/OctopusSolutionsEngineering/OctoAISpaceBuilder/internal/domain/environment"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/lumberjack.v2"
)

func ConfigureZapLogger() {
	// Create a custom encoder configuration for plain text logging
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Use ISO8601 for timestamps
	encoderConfig.LevelKey = "level"
	encoderConfig.MessageKey = "message"

	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// Always write to stdout
	stdoutCore := zapcore.NewCore(
		encoder,
		zapcore.Lock(os.Stdout),
		zapcore.InfoLevel,
	)

	core := zapcore.Core(stdoutCore)

	// Optionally tee to a rotating file when SPACEBUILDER_LOG_FILE is set
	if logFilePath := environment.GetLogFilePath(); logFilePath != "" {
		rotatingFile := &lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    100,  // megabytes before rotation
			MaxBackups: 5,    // number of old log files to keep
			MaxAge:     28,   // days to keep old log files
			Compress:   true, // gzip rotated files
		}

		fileCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(rotatingFile),
			zapcore.InfoLevel,
		)

		core = zapcore.NewTee(stdoutCore, fileCore)
	}

	// Replace the global logger with the configured logger
	logger := zap.New(core)
	zap.ReplaceGlobals(logger)
}
