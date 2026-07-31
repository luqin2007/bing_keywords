package logger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type LogEntry struct {
	Time      string   `json:"time"`
	Type      string   `json:"type,omitempty"`
	Method    string   `json:"method,omitempty"`
	Path      string   `json:"path,omitempty"`
	IP        string   `json:"ip,omitempty"`
	UserAgent string   `json:"user_agent,omitempty"`
	Count     int      `json:"count"`
	Keywords  []string `json:"keywords,omitempty"`
	Sources   []string `json:"sources,omitempty"`
	DurMs     int64    `json:"duration_ms"`
	Status    string   `json:"status,omitempty"`
	Error     string   `json:"error,omitempty"`
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

func (l *Logger) FilePath() string {
	return l.filePath
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

type LogPage struct {
	Entries []LogEntry `json:"entries"`
	Total   int        `json:"total"`
	Page    int        `json:"page"`
	Size    int        `json:"size"`
}

func (l *Logger) ReadLogs(page, size int, statusFilter, typeFilter string) (*LogPage, error) {
	if !l.enabled || l.filePath == "" {
		return &LogPage{Entries: []LogEntry{}, Total: 0, Page: page, Size: size}, nil
	}

	f, err := os.Open(l.filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var allEntries []LogEntry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 512*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry LogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		if statusFilter != "" && entry.Status != statusFilter {
			continue
		}
		if typeFilter != "" && entry.Type != typeFilter {
			continue
		}
		allEntries = append(allEntries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	total := len(allEntries)

	start := (page - 1) * size
	if start >= total {
		return &LogPage{Entries: []LogEntry{}, Total: total, Page: page, Size: size}, nil
	}
	end := start + size
	if end > total {
		end = total
	}

	entries := make([]LogEntry, 0, end-start)
	for i := start; i < end; i++ {
		entries = append(entries, allEntries[total-1-i])
	}

	return &LogPage{Entries: entries, Total: total, Page: page, Size: size}, nil
}