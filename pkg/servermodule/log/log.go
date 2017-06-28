package log

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"os"

	"runtime"

	log "github.com/sirupsen/logrus"
)

// Log 로그 설정
type Log struct {
	Path       string `json:"path"`
	Level      string `json:"level"`
	Formatter  string `json:"formatter"`
	Console    bool   `json:"console"`
	WithModule bool   `json:"with_module"`
	WithFunc   bool   `json:"with_func"`
	WithLine   bool   `json:"with_line"`
}

type Logger struct {
	logger *log.Logger
	conf   *Log
}

var logger *Logger
var once sync.Once

func GetLogger(params ...interface{}) (*log.Logger, error) {
	once.Do(func() {

		log.SetFormatter(&log.JSONFormatter{})

		logger = &Logger{}
		logger.logger = log.StandardLogger()

		logger.conf = params[0].(*Log)

		// create a log directory
		dir, _ := filepath.Split(logger.conf.Path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Error(err)
			}
		}

		// open a log file
		f, err := os.OpenFile(logger.conf.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
		if err != nil {
			log.Error("Log path is [", logger.conf.Path, "],", err)
		}
		if logger.conf.Console {
			log.SetOutput(io.MultiWriter(f, os.Stdout))
		} else {
			log.SetOutput(f)
		}

		level, err := log.ParseLevel(logger.conf.Level)
		if err != nil {
			level = log.InfoLevel
		}
		log.SetLevel(level)

		if logger.conf.Formatter == "json" {
			log.SetFormatter(&log.JSONFormatter{})
		} else {
			log.SetFormatter(&log.TextFormatter{})
		}
	})

	if len(params) == 0 {
		if logger == nil {
			return nil, errors.New("Logger is not initialized")
		}
		return logger.logger, nil
	}

	return logger.logger, nil
}

func getLoggerWithRuntimeContext() *log.Entry {

	if pc, file, line, ok := runtime.Caller(2); ok {
		funname := runtime.FuncForPC(pc).Name()

		logFields := log.Fields{}
		if logger.conf.WithModule {
			logFields["file"] = file
		}
		if logger.conf.WithFunc {
			logFields["func"] = funname
		}
		if logger.conf.WithLine {
			logFields["line"] = line
		}

		return logger.logger.WithFields(logFields)
	} else {
		return nil
	}
}

func Debug(args ...interface{}) {
	if logger == nil {
		fmt.Println(args)
		return
	}
	if entry := getLoggerWithRuntimeContext(); entry != nil {
		entry.Debug(args)
	} else {
		logger.logger.Debug(args)
	}
}

func Info(args ...interface{}) {
	if logger == nil {
		fmt.Println(args)
		return
	}
	if entry := getLoggerWithRuntimeContext(); entry != nil {
		entry.Info(args)
	} else {
		logger.logger.Info(args)
	}
}

func Warn(args ...interface{}) {
	if logger == nil {
		fmt.Println(args)
		return
	}
	if entry := getLoggerWithRuntimeContext(); entry != nil {
		entry.Warn(args)
	} else {
		logger.logger.Warn(args)
	}
}

func Error(args ...interface{}) {
	if logger == nil {
		fmt.Println(args)
		return
	}
	if entry := getLoggerWithRuntimeContext(); entry != nil {
		entry.Error(args)
	} else {
		logger.logger.Error(args)
	}
}

func Fatal(args ...interface{}) {
	if logger == nil {
		fmt.Println(args)
		return
	}
	if entry := getLoggerWithRuntimeContext(); entry != nil {
		entry.Fatal(args)
	} else {
		logger.logger.Fatal(args)
	}
}

func Panic(args ...interface{}) {
	if logger == nil {
		fmt.Println(args)
		return
	}
	if entry := getLoggerWithRuntimeContext(); entry != nil {
		entry.Panic(args)
	} else {
		logger.logger.Panic(args)
	}
}

func Debugf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Printf(format, args)
		return
	}
	if entry := getLoggerWithRuntimeContext(); entry != nil {
		entry.Debugf(format, args)
	} else {
		logger.logger.Debugf(format, args)
	}
}

func Infof(format string, args ...interface{}) {
	if logger == nil {
		fmt.Printf(format, args)
		return
	}
	if entry := getLoggerWithRuntimeContext(); entry != nil {
		entry.Infof(format, args)
	} else {
		logger.logger.Infof(format, args)
	}
}

func Warnf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Printf(format, args)
		return
	}
	if entry := getLoggerWithRuntimeContext(); entry != nil {
		entry.Warnf(format, args)
	} else {
		logger.logger.Warnf(format, args)
	}
}

func Errorf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Errorf(format, args)
		return
	}
	if entry := getLoggerWithRuntimeContext(); entry != nil {
		entry.Errorf(format, args)
	} else {
		logger.logger.Errorf(format, args)
	}
}

func Fatalf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Errorf(format, args)
		return
	}
	if entry := getLoggerWithRuntimeContext(); entry != nil {
		entry.Fatalf(format, args)
	} else {
		logger.logger.Fatalf(format, args)
	}
}

func Panicf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Errorf(format, args)
	}
	if entry := getLoggerWithRuntimeContext(); entry != nil {
		entry.Panicf(format, args)
	} else {
		logger.logger.Panicf(format, args)
	}
}
