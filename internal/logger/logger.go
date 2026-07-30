package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type LogEntry struct {
	Time     string   `json:"time"`
	Method   string   `json:"method"`
	Path     string   `json:"path"`
	Count    int      `json:"count"`
	Keywords []string `json:"keywords"`
	DurMs    int64    `json:"duration_ms"`
	Status   string   `json:"status"`
	Error    string   `json:"error"`
}

type Logger struct {
	mu       sync.Mutex
	file     *os.File
	filePath string
	enabled  bool
}

func New(filePath string) *Logger {
	l := &Logger{filePath: filePath, enabled: filePath != ""}
	if !l.enabled {
		return l
	}
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("[logger] cannot open log file %s: %v, fallback to stdout only", filePath, err)
		l.enabled = false
		return l
	}
	l.file = f
	return l
}

func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

func (l *Logger) Write(entry LogEntry) {
	entry.Time = time.Now().Format(time.RFC3339)
	data, _ := json.Marshal(entry)
	line := string(data)

	fmt.Println(line)

	if l.enabled && l.file != nil {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.file.WriteString(line + "\n")
	}
}