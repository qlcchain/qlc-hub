package log

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/qlcchain/qlc-hub/pkg/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logfile = "hub.log"
)

var (
	logger *zap.Logger
	Root   *zap.SugaredLogger
)

func init() {
	defaultLogger, _ := zap.NewDevelopment()
	logger = defaultLogger
	Root = defaultLogger.Sugar().Named("log")
}

func Setup(dir, level string) (err error) {
	err = util.CreateDirIfNotExist(dir)
	if err != nil {
		return
	}
	logfile, _ := filepath.Abs(filepath.Join(dir, logfile))
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logfile,
		MaxSize:    10, // megabytes
		MaxBackups: 10,
		MaxAge:     28, // days
		Compress:   true,
		LocalTime:  true,
	})
	l := zap.ErrorLevel
	if err := l.Set(level); err != nil {
		fmt.Println(err)
	}
	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleDebugging, l),
		zapcore.NewCore(zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}), w, l),
	)

	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	return nil
}

func Teardown() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}

//NewLogger create logger by name
func NewLogger(name string) *zap.SugaredLogger {
	return logger.Sugar().Named(name)
}
