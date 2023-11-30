package logger

import (
	"log"
	"os"
	"reflect"
	"runtime"
)

var ( // TODO delete
	Info    = log.New(os.Stdout, "INFO - App: ", log.Ldate|log.Ltime).Printf
	Warning = log.New(os.Stdout, "WARNING - App: ", log.Ldate|log.Ltime|log.Lshortfile).Printf
	Error   = log.New(os.Stderr, "ERROR - App: ", log.Ldate|log.Ltime|log.Lshortfile).Println
)

func CreateLoggers() (errLog, infoLog *log.Logger) {
	errLog = log.New(os.Stderr, "ERROR - App: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile) // Creates logs of errors
	infoLogFile, err := os.OpenFile("info.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o664)
	if err != nil {
		errLog.Printf("Cannot open a log file. Error is %s\nStdout will be used for the info log ", err)
		infoLogFile = os.Stdout
	}
	infoLog = log.New(infoLogFile, "INFO - App:  ", log.Ldate|log.Ltime|log.Lmicroseconds)
	return
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func GetCurrentFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}
