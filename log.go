package apollo_client

import "log"

var logger LoggerInterface = &DefaultLogger{}

type LoggerInterface interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})

	Debugf(format string, params ...interface{})
	Infof(format string, params ...interface{})
	Warnf(format string, params ...interface{})
	Errorf(format string, params ...interface{})
}

func SetLogger(l LoggerInterface) {
	if logger != nil {
		logger = l
	}
}

type DefaultLogger struct{}

func (l *DefaultLogger) Debug(v ...interface{}) {
	log.Print(v...)
}

func (l *DefaultLogger) Info(v ...interface{}) {
	log.Print(v...)
}

func (l *DefaultLogger) Warn(v ...interface{}) {
	log.Print(v...)
}

func (l *DefaultLogger) Error(v ...interface{}) {
	log.Print(v...)
}

func (l *DefaultLogger) Debugf(format string, params ...interface{}) {
	log.Printf(format, params...)
}

func (l *DefaultLogger) Infof(format string, params ...interface{}) {
	log.Printf(format, params...)
}

func (l *DefaultLogger) Warnf(format string, params ...interface{}) {
	log.Printf(format, params...)
}

func (l *DefaultLogger) Errorf(format string, params ...interface{}) {
	log.Printf(format, params...)
}
