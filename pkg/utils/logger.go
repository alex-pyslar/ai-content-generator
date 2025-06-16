// pkg/utils/logger.go
package utils

import (
	"log"
	"os"
	"sync"
)

// Logger предоставляет обертку вокруг стандартного log.Logger
// для унифицированного логирования.
type Logger struct {
	*log.Logger
}

var (
	loggerInstance *Logger
	once           sync.Once
)

// NewLogger создает или возвращает единственный экземпляр логгера.
func NewLogger() *Logger {
	once.Do(func() {
		loggerInstance = &Logger{
			Logger: log.New(os.Stdout, "[AI-GEN] ", log.Ldate|log.Ltime|log.Lshortfile),
		}
	})
	return loggerInstance
}

// Info логирует информационное сообщение.
func (l *Logger) Info(format string, v ...interface{}) {
	l.Printf("INFO: "+format, v...)
}

// Warn логирует предупреждение.
func (l *Logger) Warn(format string, v ...interface{}) {
	l.Printf("WARN: "+format, v...)
}

// Error логирует ошибку.
func (l *Logger) Error(format string, v ...interface{}) {
	l.Printf("ERROR: "+format, v...)
}

// Fatal логирует фатальную ошибку и завершает программу.
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.Fatalf("FATAL: "+format, v...)
}
