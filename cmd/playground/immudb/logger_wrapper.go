package main

import (
	"fmt"

	immulogger "github.com/codenotary/immudb/pkg/logger"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

type immuLoggerWrapper struct {
	logger logging.Logger
}

func wrapLogger(logger logging.Logger) *immuLoggerWrapper {
	return &immuLoggerWrapper{logger: logger}
}

func (l *immuLoggerWrapper) Errorf(msg string, args ...interface{}) {
	l.logger.Error(nil, fmt.Sprintf(msg, args...))
}

func (l *immuLoggerWrapper) Warningf(msg string, args ...interface{}) {
	l.logger.Debug(fmt.Sprintf(msg, args...))
}

func (l *immuLoggerWrapper) Infof(msg string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}

func (l *immuLoggerWrapper) Debugf(msg string, args ...interface{}) {
	l.logger.Debug(fmt.Sprintf(msg, args...))
}

func (l *immuLoggerWrapper) CloneWithLevel(level immulogger.LogLevel) immulogger.Logger {
	switch level {
	case immulogger.LogDebug:
		l.logger.SetLevel(logging.DebugLevel)
	case immulogger.LogInfo:
		l.logger.SetLevel(logging.InfoLevel)
	case immulogger.LogWarn:
		l.logger.SetLevel(logging.WarnLevel)
	case immulogger.LogError:
		l.logger.SetLevel(logging.ErrorLevel)
	}

	return l
}
