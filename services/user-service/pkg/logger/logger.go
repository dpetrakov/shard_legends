package logger

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Logger struct {
	logger *log.Logger
}

type LogEntry struct {
	Level     string                 `json:"level"`
	Timestamp string                 `json:"timestamp"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

func New() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", 0),
	}
}

func (l *Logger) logWithLevel(level, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Level:     level,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Message:   message,
		Fields:    fields,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		l.logger.Printf("Failed to marshal log entry: %v", err)
		return
	}

	l.logger.Println(string(jsonData))
}

func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.logWithLevel("info", message, f)
}

func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.logWithLevel("error", message, f)
}

func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.logWithLevel("warn", message, f)
}

func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.logWithLevel("debug", message, f)
}