package logger

import (
	"fmt"
	"time"
)

type Logger struct{}

// const (
// 	info = iota + 1
// 	success
// 	warning
// 	err
// )

const (
	reset   = "\033[0m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	gray    = "\033[37m"
	white   = "\033[97m"
)

func (l Logger) print(color string, _type string, data ...any) {
	fmt.Printf(color+"%v. %v: %v.\n"+reset, time.Now().Format(time.Layout), _type, data)
}

func (l Logger) Success(data ...any) {
	l.print(green, "Success", data...)
}

func (l Logger) Info(data ...any) {
	l.print(cyan, "Info", data...)
}

func (l Logger) Warning(data ...any) {
	l.print(yellow, "Warning", data...)
}

func (l Logger) Error(data ...any) {
	l.print(red, "Error", data...)
}
