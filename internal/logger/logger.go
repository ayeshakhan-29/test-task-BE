package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// Log levels
const (
	DebugLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

var (
	logger *log.Logger
	level  int
)

// InitLogger initializes the logger with the specified log level
func InitLogger() {
	// Default to Info level
	level = InfoLevel

	// Create logs directory if it doesn't exist
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.Mkdir(logDir, 0755); err != nil {
			log.Fatalf("Failed to create logs directory: %v", err)
		}
	}

	// Create or open the log file
	logFile := filepath.Join(logDir, "app.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Initialize logger with file and stdout output
	multiWriter := io.MultiWriter(os.Stdout, file)
	logger = log.New(multiWriter, "", log.LstdFlags|log.Lmicroseconds)
}

// SetLevel sets the log level
func SetLevel(lvl int) {
	level = lvl
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	if level <= DebugLevel {
		logWithCaller("DEBUG", format, v...)
	}
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	if level <= InfoLevel {
		logWithCaller("INFO ", format, v...)
	}
}

// Warn logs a warning message
func Warn(format string, v ...interface{}) {
	if level <= WarnLevel {
		logWithCaller("WARN ", format, v...)
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	if level <= ErrorLevel {
		logWithCaller("ERROR", format, v...)
	}
}

// Fatal logs a fatal error and exits
func Fatal(format string, v ...interface{}) {
	logWithCaller("FATAL", format, v...)
	os.Exit(1)
}

// Fatalf logs a fatal error with format and exits
func Fatalf(format string, v ...interface{}) {
	logWithCaller("FATAL", format, v...)
	os.Exit(1)
}

// logWithCaller adds the caller's file and line number to the log message
func logWithCaller(level, format string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(2) // Adjust the call depth as needed
	if !ok {
		file = "???"
		line = 0
	}
	// Get just the filename from the full path
	file = filepath.Base(file)

	// Format the message with the caller info
	msg := fmt.Sprintf("[%s] %s:%d %s", level, file, line, format)
	if len(v) > 0 {
		msg = fmt.Sprintf(msg, v...)
	}

	// Log the message
	logger.Output(2, msg) // Use depth 2 to get the correct caller info
}
