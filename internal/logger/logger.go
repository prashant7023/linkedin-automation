package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger represents a structured logger
type Logger struct {
	level string
}

// New creates a new logger instance
func New(level string) *Logger {
	log.SetFlags(0) // Remove default timestamp since we'll add our own
	return &Logger{level: level}
}

// formatMessage formats a log message with timestamp and level
func (l *Logger) formatMessage(level, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("[%s] [%s] %s", timestamp, level, message)
}

// Debug logs debug-level messages
func (l *Logger) Debug(message string) {
	if l.level == "debug" {
		fmt.Println(l.formatMessage("DEBUG", message))
	}
}

// Info logs info-level messages
func (l *Logger) Info(message string) {
	if l.level == "debug" || l.level == "info" {
		fmt.Println(l.formatMessage("INFO", message))
	}
}

// Warn logs warning-level messages
func (l *Logger) Warn(message string) {
	if l.level != "error" {
		fmt.Println(l.formatMessage("WARN", message))
	}
}

// Error logs error-level messages
func (l *Logger) Error(message string) {
	fmt.Fprintf(os.Stderr, "%s\n", l.formatMessage("ERROR", message))
}

// Infof logs formatted info messages
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// Errorf logs formatted error messages
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

// Warnf logs formatted warning messages
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

// Debugf logs formatted debug messages
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}
