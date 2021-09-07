package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjackv2 "gopkg.in/natefinch/lumberjack.v2"
)

const (
	Console = "console"
	File    = "file"
)

var (
	Level = zap.DebugLevel
	Target = Console
	MaxSize = 5
	MaxBackups = 30
	MaxAge = 7
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func Init() {
	fmt.Println("Init Logger Level --> ", Level, " Target --> ", Target, " MaxSize --> ",MaxSize, " MaxBackups -->", MaxBackups, " MaxAge --> ", MaxAge)
	w := zapcore.AddSync(&lumberjackv2.Logger{
		Filename:   "log/im.log",
		MaxSize:    MaxSize, // megabytes
		MaxBackups: MaxBackups,
		MaxAge:     MaxAge, // days
	})

	var writeSyncer zapcore.WriteSyncer
	if Target == Console {
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))
	}
	if Target == File {
		writeSyncer = zapcore.NewMultiWriteSyncer(w)
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(NewEncoderConfig()),
		writeSyncer,
		Level,
	)
	Logger = zap.New(core, zap.AddCaller())
	Sugar = Logger.Sugar()
}
