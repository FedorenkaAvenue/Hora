package logger

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	info    = "info"
	success = "success"
	warning = "warning"
	error   = "error"
)

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

type Logger struct{}

func (l Logger) Info(data ...any) {
	l.print(cyan, info, data...)
	l.writeLog(info, data...)
}

func (l Logger) Success(data ...any) {
	l.print(green, success, data...)
}

func (l Logger) Warning(data ...any) {
	l.print(yellow, warning, data...)
	l.writeLog(warning, data...)
}

func (l Logger) Error(data ...any) {
	l.print(red, error, data...)
	l.writeLog(error, data...)
}

func (l Logger) print(color string, _type string, data ...any) {
	fmt.Printf(color+"%v. %v: %v.\n"+reset, time.Now().Format(time.Layout), strings.ToUpper(_type), data)
}

func (l Logger) writeLog(_type string, data ...any) {
	f, err := os.OpenFile(fmt.Sprintf("%v.log", _type), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("%v. %v.\n", time.Now().Format(time.Layout), data)); err != nil {
		panic(err)
	}
}
