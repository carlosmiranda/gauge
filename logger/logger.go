// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/op/go-logging"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	LOGS_DIRECTORY   = "logs_directory"
	logs             = "logs"
	gaugeLogFileName = "gauge.log"
	apiLogFileName   = "api.log"
)

type MyLogger interface {
	SpecStart(string)
	SpecEnd()
	ScenarioStart(string)
	ScenarioEnd(bool)
	StepStart(string)
	StepEnd(bool)
	writeSysoutBuffer(string)
	Error(string, ...interface{})
}

var currentLogger MyLogger

func Current() MyLogger {
	if currentLogger == nil {
		currentLogger = newColoredConsoleWriter()
	}
	return currentLogger
}

type GaugeLogger struct {
	*logging.Logger
}

func Info(msg string, args ...interface{}) {
	Log.Info(msg, args...)
	ConsoleWrite(msg, args...)
}

func Error(msg string, args ...interface{}) {
	Log.Error(msg, args...)
	ConsoleWrite(msg, args...)
}

func Warning(msg string, args ...interface{}) {
	Log.Warning(msg, args...)
	ConsoleWrite(msg, args...)
}

func Critical(msg string, args ...interface{}) {
	Log.Critical(msg, args...)
	ConsoleWrite(msg, args...)
}

func Debug(msg string, args ...interface{}) {
	Log.Debug(msg, args...)
	if level == logging.DEBUG {
		ConsoleWrite(msg, args...)
	}
}

func ConsoleWrite(msg string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(msg, args...))
}

func (g GaugeLogger) Write(b []byte) (int, error) {
	str := strings.Trim(string(b), "\n ")
	if len(str) > 0 {
		Current().writeSysoutBuffer(fmt.Sprintf("\t%s", str))
	}
	return len(b), nil
}

var Log = GaugeLogger{logging.MustGetLogger("gauge")}
var ApiLog = GaugeLogger{logging.MustGetLogger("gauge-api")}

var gaugeLogFile = filepath.Join(logs, gaugeLogFileName)
var apiLogFile = filepath.Join(logs, apiLogFileName)
var level logging.Level

var coloredFormat = logging.MustStringFormatter(
	"%{color}%{message}%{color:reset}",
)

var uncoloredFormat = logging.MustStringFormatter(
	"%{time:15:04:05.000} %{message}",
)

func Initialize(verbose bool, logLevel string, simpleConsoleOutput bool) {
	level = loggingLevel(verbose, logLevel)
	initFileLogger(level, simpleConsoleOutput)
	initApiLogger(level, simpleConsoleOutput)
}

func getLogFormatter(logger logging.Backend, supportsColoredFormat bool, simpleConsoleOutput bool) logging.Backend {
	if supportsColoredFormat && !simpleConsoleOutput {
		return logging.NewBackendFormatter(logger, coloredFormat)
	}
	return logging.NewBackendFormatter(logger, uncoloredFormat)
}

func initFileLogger(level logging.Level, simpleConsoleOutput bool) {
	logsDir := os.Getenv(LOGS_DIRECTORY)
	var gaugeFileLogger logging.Backend
	if logsDir == "" {
		gaugeFileLogger = createFileLogger(gaugeLogFile, 20)
	} else {
		gaugeFileLogger = createFileLogger(filepath.Join(logsDir, gaugeLogFileName), 20)
	}
	fileFormatter := getLogFormatter(gaugeFileLogger, false, simpleConsoleOutput)

	fileLoggerLeveled := logging.AddModuleLevel(fileFormatter)
	fileLoggerLeveled.SetLevel(logging.DEBUG, "")

	logging.SetBackend(fileLoggerLeveled)
}

func initApiLogger(level logging.Level, simpleConsoleOutput bool) {
	logsDir, err := filepath.Abs(os.Getenv(LOGS_DIRECTORY))
	var apiFileLogger logging.Backend
	if logsDir == "" || err != nil {
		apiFileLogger = createFileLogger(apiLogFile, 10)
	} else {
		apiFileLogger = createFileLogger(filepath.Join(logsDir, apiLogFileName), 10)
	}

	fileFormatter := getLogFormatter(apiFileLogger, false, simpleConsoleOutput)
	fileLoggerLeveled := logging.AddModuleLevel(fileFormatter)
	fileLoggerLeveled.SetLevel(level, "")
	ApiLog.SetBackend(fileLoggerLeveled)
}

func NewParallelLogger(n int) *GaugeLogger {
	parallelLogger := &GaugeLogger{logging.MustGetLogger("gauge")}
	stdOutLogger := logging.NewLogBackend(os.Stdout, "", 0)
	stdOutFormatter := logging.NewBackendFormatter(stdOutLogger, logging.MustStringFormatter("[runner:"+strconv.Itoa(n)+"] %{message}"))
	stdOutLoggerLeveled := logging.AddModuleLevel(stdOutFormatter)
	stdOutLoggerLeveled.SetLevel(level, "")
	parallelLogger.SetBackend(stdOutLoggerLeveled)
	return parallelLogger
}

func createFileLogger(name string, size int) logging.Backend {
	if !filepath.IsAbs(name) {
		name = getLogFile(name)
	}
	return logging.NewLogBackend(&lumberjack.Logger{
		Filename:   name,
		MaxSize:    size, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
	}, "", 0)
}

func getLogFile(fileName string) string {
	if config.ProjectRoot != "" {
		return filepath.Join(config.ProjectRoot, fileName)
	} else {
		gaugeHome, err := common.GetGaugeHomeDirectory()
		if err != nil {
			return fileName
		}
		return filepath.Join(gaugeHome, fileName)
	}
}

func loggingLevel(verbose bool, logLevel string) logging.Level {
	if verbose {
		return logging.DEBUG
	}
	if logLevel != "" {
		switch strings.ToLower(logLevel) {
		case "debug":
			return logging.DEBUG
		case "info":
			return logging.INFO
		case "warning":
			return logging.WARNING
		case "error":
			return logging.ERROR
		case "critical":
			return logging.CRITICAL
		case "notice":
			return logging.NOTICE
		}
	}
	return logging.INFO
}

func HandleWarningMessages(warnings []string) {
	for _, warning := range warnings {
		Log.Warning(warning)
	}
}
