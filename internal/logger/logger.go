package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
)

type level string
type loggerLevel int

const (
	defaultTimeFormat = "2006-01-02T15:04:05.000000Z07:00"

	debug   level = "DEBUG"
	info    level = "INFO"
	warning level = "WARN"
	error   level = "ERROR"
	fatal   level = "FATAL"
)

const (
	loggerLevelDebug loggerLevel = iota
	loggerLevelInfo
	loggerLevelWarning
	loggerLevelError
	loggerLevelFatal
	loggerLevelDisable
)

type loggerStack interface {
	Debug(a ...any)
	Info(a ...any)
	Warn(a ...any)
	Error(a ...any)
	Fatal(a ...any)
	Print(a ...any) //prints the stack
	Get() string    //get the stack
}

type logStack struct {
	stack []string
}

func NewLoggerStack() loggerStack {
	return &logStack{}
}

func (ls *logStack) Debug(a ...any) {
	if getLoggerLevel() <= loggerLevelDebug {
		ls.stack = append(ls.stack, generateLog(debug, a...))
	}
}

func (ls *logStack) Info(a ...any) {
	if getLoggerLevel() <= loggerLevelInfo {
		ls.stack = append(ls.stack, generateLog(info, a...))
	}
}

func (ls *logStack) Warn(a ...any) {
	if getLoggerLevel() <= loggerLevelWarning {
		ls.stack = append(ls.stack, generateLog(warning, a...))
	}
}

func (ls *logStack) Error(a ...any) {
	if getLoggerLevel() <= loggerLevelError {
		ls.stack = append(ls.stack, generateLog(error, a...))
	}
}

func (ls *logStack) Fatal(a ...any) {
	if getLoggerLevel() <= loggerLevelFatal {
		ls.stack = append(ls.stack, generateLog(fatal, a...))
		os.Exit(1)
	}
}

func (ls *logStack) Print(a ...any) {
	defer ls.cleanup()
	if settings.GetEnvironment().String() == "local" {
		fmt.Printf("%v[%v]: %v%v\n", time.Now().Format(timeFormat), setTextColor("STACK", colorMagenta), fmt.Sprint(a...), ls.Get())
	} else {
		fmt.Printf("%v[%v]: %v%v\n", time.Now().Format(timeFormat), "STACK", fmt.Sprint(a...), ls.Get())
	}
}

func (ls *logStack) Get() string {
	defer ls.cleanup()
	return jsonString(ls.stack)
}

func (ls *logStack) cleanup() {
	ls.stack = nil
}

func generateLog(l level, a ...any) string {
	return fmt.Sprintf("%v[%v]: %v\n", time.Now().Format(timeFormat), l, fmt.Sprint(a...))
}

//static functions
var (
	isColorEnabled bool = true
	timeFormat          = defaultTimeFormat
)

func ColorEnable(enable bool) {
	isColorEnabled = enable
}

func SetTimeFormal(f string) {
	timeFormat = f
}

func Debug(a ...any) bool {
	if getLoggerLevel() > loggerLevelDebug {
		return false
	}
	return logBase(debug, a...)
}

func Info(a ...any) bool {
	if getLoggerLevel() > loggerLevelInfo {
		return false
	}
	return logBase(info, a...)
}

func Warning(a ...any) bool {
	if getLoggerLevel() > loggerLevelWarning {
		return false
	}
	return logBase(warning, a...)
}

func Error(a ...any) bool {
	if getLoggerLevel() > loggerLevelError {
		return false
	}
	return logBase(error, a...)
}

func Fatal(a ...any) {
	if getLoggerLevel() > loggerLevelFatal {
		return
	}
	logBase(fatal, a...)
	os.Exit(1)
}

func logBase(l level, a ...any) bool {
	if settings.GetEnvironment().String() == "local" {
		fmt.Printf("%v[%v]: %v\n", time.Now().Format(timeFormat), logColorLevel(l), fmt.Sprint(a...))
	} else {
		fmt.Printf("%v[%v]: %v\n", time.Now().Format(timeFormat), l, fmt.Sprint(a...))
	}
	return true
}

func getLoggerLevel() loggerLevel {
	return loggerLevel(settings.GetLoggerLevel().Int())
}

func logColorLevel(l level) string {
	text := string(l)

	if isColorEnabled {

		switch l {
		case debug:
			return setTextColor(text, colorCyan)
		case info:
			return setTextColor(text, colorGreen)
		case warning:
			return setTextColor(text, colorYellow)
		case error:
			return setTextColor(text, colorRed)
		case fatal:
			return setTextColor(text, colorMagenta)
		}
	}
	return text
}

func jsonString(a any) string {
	jsonStr, _ := json.Marshal(a)

	return string(jsonStr)
}
