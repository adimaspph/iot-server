package config

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewLogger(viper *viper.Viper) *logrus.Logger {
	log := logrus.New()

	// Set log level from config
	log.SetLevel(logrus.Level(viper.GetInt("LOG_LEVEL")))

	// Output to stdout (best for Docker/K8s)
	log.SetOutput(os.Stdout)

	// Enable caller reporting
	log.SetReportCaller(true)

	// Custom formatter with UTC timestamps
	log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat:  time.RFC3339Nano, // e.g. 2025-08-24T19:45:12.123456Z
		FullTimestamp:    true,
		ForceColors:      true,
		PadLevelText:     true,
		QuoteEmptyFields: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			// Shorten function name
			funcName := f.Function
			if idx := strings.LastIndex(funcName, "/"); idx != -1 {
				funcName = funcName[idx+1:]
			}

			// Shorten file path
			fileName := f.File
			if idx := strings.LastIndex(fileName, "/"); idx != -1 {
				fileName = fileName[idx+1:]
			}

			// Properly format line number
			return funcName, fileName + ":" + strconv.Itoa(f.Line)
		},
	})

	// Hook to ensure log timestamps are always in UTC
	log.AddHook(&utcHook{})

	return log
}

// utcHook forces timestamps to UTC
type utcHook struct{}

func (h *utcHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *utcHook) Fire(entry *logrus.Entry) error {
	entry.Time = entry.Time.UTC()
	return nil
}
