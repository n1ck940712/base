package logger

import (
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	os.Setenv("Test", "123")

	Debug("123: ", os.Getenv("Test"))
	os.Setenv("Test", "453")
	Debug("123: ", os.Getenv("Test"))
	Debug("Debug")
	Info("Info")
	Warning("Warning")
	Error("Error")
	Debug("Debug1")
	Fatal("FATAL")
}
