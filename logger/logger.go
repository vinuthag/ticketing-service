//    Copyright (c) 2019 by General Electric Company. All rights reserved.
//
//      The copyright to the computer software herein is the property of
//      General Electric Company. The software may be used and/or copied only
//      with the written permission of General Electric Company or in accordance
//      with the terms and conditions stipulated in the agreement/contract
//      under which the software has been supplied.

package logger

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	util "ticketing-service/util"

	log "github.com/sirupsen/logrus"
)

const (
	TRACE = "trace"
	DEBUG = "debug"
	INFO  = "info"
	WARN  = "warn"
	ERROR = "error"
	FATAL = "fatal"
)

// Init function loads the log file or creates if already not present
// log is set with the formatter (it can be TextFormatter or JSON formatter),
// mode of output and log level

func init() {
	log.SetFormatter(&log.JSONFormatter{FieldMap: log.FieldMap{
		"msg": "message",
	}}) // log format is of JSON type
	log.SetLevel(logLevel(util.GetProperty(util.LOG_LEVEL))) // To set log level
}

// This function is to append the log with the fileds 'file'
// and 'function' which are file_name with line number
// and the function name respectively of the logger statement
// log example : {"file":"main.go:17","function":"main","level":"info",
// "msg":"Started preference service","time":"2020-02-14T11:24:23+05:30"}

func Log() *log.Entry {

	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("Could not get context info for logger!")
		return log.WithField("file", "").WithField("function", "")
	}
	filename := file[strings.LastIndex(file, "/")+1:] + ":" + strconv.Itoa(line)
	funcname := runtime.FuncForPC(pc).Name()
	fn := funcname[strings.LastIndex(funcname, ".")+1:]
	return log.WithField("file", filename).WithField("function", fn)

}

// This function is to set the logger level
func logLevel(level string) log.Level {
	switch level {
	case TRACE:
		return log.TraceLevel
	case DEBUG:
		return log.DebugLevel
	case INFO:
		return log.InfoLevel
	case WARN:
		return log.WarnLevel
	case ERROR:
		return log.ErrorLevel
	case FATAL:
		return log.FatalLevel
	default:
		return log.InfoLevel
	}
}
