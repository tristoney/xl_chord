package log

import (
	"fmt"
	"log"
)

type Level int32

const (
	INFO Level = 1
	DEBUG Level = 2
	ERROR Level = 3
)

func (l Level) String() string {
	switch l {
	case INFO:
		return fmt.Sprintf("[INFO] ")
	case DEBUG:
		return fmt.Sprintf("[DEBUG] ")
	case ERROR:
		return fmt.Sprintf("[ERROR] ")
	default:
		return fmt.Sprintf("[INFO] ")
	}

}

func Logf(level Level, format string, v ...interface{}) {
	format = fmt.Sprintf("%s", level) + format
	log.Printf(format, v...)
}

func Logln(level Level, str string) {
	log.Println(fmt.Sprintf("%s", level) + str)
}
