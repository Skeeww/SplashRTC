package main

import "fmt"

type LogLevel int

const (
	Trace LogLevel = iota
	Debug
	Info
	Warn
	Error
)

type Loggerer interface {
	Trace(args ...any)
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Log(level LogLevel, args ...any)
}

type Logger struct {
	Name  string
	Level LogLevel
}

type LoggerOptions struct {
	Level LogLevel
}

func CreateLogger(name string, opts *LoggerOptions) Loggerer {
	logger := &Logger{
		Name:  name,
		Level: Info,
	}

	if opts != nil {
		logger.Level = opts.Level
	}

	return logger
}

func (l LogLevel) String() string {
	switch l {
	case Trace:
		return "TRACE"
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	default:
		return "LOG"
	}
}

func (l *Logger) Log(level LogLevel, args ...any) {
	if l.Level > level {
		return
	}
	fmt.Printf("(%s)\t[%s]\t%v\n", l.Name, level.String(), args)
}

func (l *Logger) Trace(args ...any) {
	l.Log(Trace, args...)
}

func (l *Logger) Debug(args ...any) {
	l.Log(Debug, args...)
}

func (l *Logger) Info(args ...any) {
	l.Log(Info, args...)
}

func (l *Logger) Warn(args ...any) {
	l.Log(Warn, args...)
}

func (l *Logger) Error(args ...any) {
	l.Log(Error, args...)
}
