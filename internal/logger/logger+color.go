package logger

import "fmt"

type logColor string

const (
	colorReset   logColor = "\u001b[0m"
	colorCyan    logColor = "\u001b[36;1m"
	colorGreen   logColor = "\u001b[32;1m"
	colorYellow  logColor = "\u001b[33;1m"
	colorRed     logColor = "\u001b[31;1m"
	colorMagenta logColor = "\u001b[35;1m"
)

func setTextColor(text string, color logColor) string {
	return fmt.Sprintf("%v%v%v", color, text, colorReset)
}
