package logging

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ContextKey string

const (
	ProjectNameKey ContextKey = "project_name"
	WorkerIdKey    ContextKey = "worker_id"
)

func InitLogging(logFilePath string) error {
	logrus.SetOutput(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    1,
		MaxBackups: 1,
		MaxAge:     7,
		Compress:   true,
	})
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.AddHook(&ContextHook{})

	return nil
}

type ContextHook struct{}

func (h *ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *ContextHook) Fire(entry *logrus.Entry) error {
	if entry.Context == nil {
		return nil
	}

	if projectName, ok := entry.Context.Value(ProjectNameKey).(string); ok {
		entry.Data[string(ProjectNameKey)] = projectName
	}

	if workerId, ok := entry.Context.Value(WorkerIdKey).(int); ok {
		entry.Data[string(WorkerIdKey)] = workerId
	}

	return nil
}
