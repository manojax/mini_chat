package utils

import (
	"context"
	"fmt"
	"time"
)

type LogLevel int

var (
	LL = NewCLogger()
)

const (
	DEBUG LogLevel = iota + 1
	INFO
	WARN
	ERROR
)

type LogMessage struct {
	Level   LogLevel
	Time    time.Time
	Message string
}

func (s *LogMessage) ToString() string {
	switch s.Level {
	case DEBUG:
		return fmt.Sprintf("%s: [blue]DEBUG [white]%s", s.Time.Format(time.RFC3339), s.Message)
	case INFO:
		return fmt.Sprintf("%s: [green]INFO [white]%s", s.Time.Format(time.RFC3339), s.Message)
	case WARN:
		return fmt.Sprintf("%s: [yellow]WARN [white]%s", s.Time.Format(time.RFC3339), s.Message)
	case ERROR:
		return fmt.Sprintf("%s: [red]ERROR [white]%s", s.Time.Format(time.RFC3339), s.Message)
	}
	return fmt.Sprintf("%s: [red]UNKNOW [white]%s", s.Time.Format(time.RFC3339), s.Message)
}

type CLogger struct {
	notification chan *LogMessage
}

func NewCLogger() *CLogger {
	return &CLogger{
		notification: make(chan *LogMessage, 10),
	}
}

func (s *CLogger) Exec(ctx context.Context, fn func(string)) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-s.notification:
			if !ok {
				return
			}
			fn(msg.ToString())
		}
	}
}

func (s *CLogger) Debug(format string, a ...any) {
	s.notification <- &LogMessage{
		Level:   DEBUG,
		Time:    time.Now(),
		Message: fmt.Sprintf(format, a...),
	}
}

func (s *CLogger) Info(format string, a ...any) {
	s.notification <- &LogMessage{
		Level:   INFO,
		Time:    time.Now(),
		Message: fmt.Sprintf(format, a...),
	}
}

func (s *CLogger) Warn(format string, a ...any) {
	s.notification <- &LogMessage{
		Level:   WARN,
		Time:    time.Now(),
		Message: fmt.Sprintf(format, a...),
	}
}

func (s *CLogger) Error(format string, a ...any) {
	s.notification <- &LogMessage{
		Level:   ERROR,
		Time:    time.Now(),
		Message: fmt.Sprintf(format, a...),
	}
}
