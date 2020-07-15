package logging

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/chongyangshi/Order/config"
)

var l = log.New(os.Stdout, "[Order] ", log.Ldate|log.Ltime)

// Log prints a standard output to stdout
func Log(format string, v ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = fmt.Sprintf("%s\n", format)
	}

	l.Printf(format, v)
}

// Fatal prints an error output to stderr and panics the program
func Fatal(format string, v ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = fmt.Sprintf("%s\n", format)
	}

	l.Fatalf(format, v)
}

// Debug prints a debug output to stdout if it is enabled in config
func Debug(format string, v ...interface{}) {
	if config.Config == nil {
		log.Println("Error: Config not loaded when Debug is called")
	}

	if !config.Config.DebugOutput {
		return
	}

	if !strings.HasSuffix(format, "\n") {
		format = fmt.Sprintf("%s\n", format)
	}

	l.Printf(format, v)
}
