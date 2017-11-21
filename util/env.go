package util

import (
	"fmt"
	"runtime"
	"strings"
)

func Message(msg string) string {
	function, file, line, _ := runtime.Caller(1)
	i := strings.LastIndex(file, "/")
	if i == -1 {
		// do nothing
	} else {
		file = file[i+1:]
	}

	if msg == "" {
		msg = "\"\""
	}

	return fmt.Sprintf("File: %s; Function: %s; Line: %d; Message: %s", file, runtime.FuncForPC(function).Name(), line, msg)
}
