package dashboard

import (
	"fmt"
	"sync"
	"time"
)

type LogBuffer struct {
	mu        sync.RWMutex
	logs      []string
	listeners []chan string
	size      int
}

func NewLogBuffer() *LogBuffer {
	return &LogBuffer{}
}

func (l *LogBuffer) AddListener() chan string {
	l.mu.Lock()
	defer l.mu.Unlock()
	ch := make(chan string, 10)
	l.listeners = append(l.listeners, ch)
	return ch
}

func (l *LogBuffer) AddLog(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	s := fmt.Sprintf(`<span class="log-time">%s </span>`, time.Now().Format("15:04:05")) + msg
	l.logs = append(l.logs, s)
	if len(l.logs) > l.size && l.size > 0 {
		l.logs = l.logs[1:]
	}
	for _, ch := range l.listeners {
		select {
		case ch <- s:
		default:
		}
	}
}

func (l *LogBuffer) GetLogs() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	logsCopy := make([]string, len(l.logs))
	copy(logsCopy, l.logs)
	return logsCopy
}
