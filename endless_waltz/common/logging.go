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

	switch strings.ToLower(LEVEL) {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	case "panic":
		logger.SetLevel(logrus.PanicLevel)
	default:
		fmt.Printf("Invalid log level: \"%s\". Using default level (INFO).\n", LEVEL)
		logger.SetLevel(logrus.InfoLevel)
	}

	// Set the log level (e.g., INFO, DEBUG, WARN, ERROR)
	return logger
}

/*
When to return certain responses:
FATAL --> Application will shut down or be unresponsive due to error (include context for error to aid troubleshooting)
ERROR --> impacts execution of specific operation within code (lower priority than fatal error)
WARN  --> something unexpected has occurred, but the application can function normally for now
INFO  --> show that the system is operating normally
DEBUG --> used for debugging lol
*/
