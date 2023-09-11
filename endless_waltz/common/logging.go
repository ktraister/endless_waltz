package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

func createLogger(LEVEL string, JSON string) *logrus.Logger {
	logger := logrus.New()
	if JSON == "JSON" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}
	// Set the log level (e.g., INFO, DEBUG, WARN, ERROR)
	setLogLevelFromString(LEVEL)
	return logger
}

func setLogLevelFromString(levelStr string) {
	switch strings.ToLower(levelStr) {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	default:
		fmt.Printf("Invalid log level: \"%s\". Using default level (INFO).\n", levelStr)
		logrus.SetLevel(logrus.InfoLevel)
	}
}

/*
When to return certain responses:
FATAL --> Application will shut down or be unresponsive due to error (include context for error to aid troubleshooting)
ERROR --> impacts execution of specific operation within code (lower priority than fatal error)
WARN  --> something unexpected has occurred, but the application can function normally for now
INFO  --> show that the system is operating normally
DEBUG --> used for debugging lol
*/
