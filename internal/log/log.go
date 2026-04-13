package log

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
)

var (
	errorColor   = color.New(color.FgRed).SprintFunc()
	successColor = color.New(color.FgGreen).SprintFunc()
	warningColor = color.New(color.FgHiYellow).SprintFunc()
)

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
}

func Fatal(v ...interface{}) {
	message := fmt.Sprint(v...)
	log.Output(2, message)
	os.Exit(1)
}

func FatalColored(v ...interface{}) {
	message := fmt.Sprint(v...)
	log.Output(2, errorColor(message))
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Output(2, message)
	os.Exit(1)
}

func FatalfColored(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Output(2, errorColor(message))
	os.Exit(1)
}

func Printf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Output(2, message)
}

func PrintfColored(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Output(2, successColor(message))
}

func Print(v ...interface{}) {
	message := fmt.Sprint(v...)
	log.Output(2, message)
}

func Println(v ...interface{}) {
	message := fmt.Sprintln(v...)
	log.Output(2, message)
}

func PrintlnColored(v ...interface{}) {
	message := fmt.Sprintln(v...)
	log.Output(2, successColor(message))
}

func Warn(v ...interface{}) {
	message := fmt.Sprint(v...)
	log.Output(2, message)
}

func Warnf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Output(2, message)
}

func WarnfColored(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Output(2, warningColor(message))
}

func Infof(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Output(2, message)
}
