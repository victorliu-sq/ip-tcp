package myDebug

import (
	"fmt"
	"io"
	"log"
	"os"
)

var Debug *log.Logger

func InitDebugger() {
	Debug = log.New(os.Stdout, "", log.Ltime|log.Lshortfile)
}

func SetDebug(enabled bool) {
	if enabled {
		Debug.SetOutput(os.Stdout)
	} else {
		Debug.SetOutput(io.Discard)
	}
}

func Debugln(formatString string, args ...interface{}) {
	Debug.Output(1, fmt.Sprintf("%v\n", fmt.Sprintf(formatString, args...)))
}
