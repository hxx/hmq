package loge

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//设置日志级别
//curl -XPUT --data '{"level":"error"}' http://localhost:5555/handle/level
var resultLogger *zap.SugaredLogger
var errorLogger *zap.SugaredLogger
var Alevel zap.AtomicLevel

func New(rootPath, serviceName string) {
	fmt.Println("rootPath----", rootPath)
	// 设置一些基本日志格式 具体含义还比较好理解，直接看zap源码也不难懂
	//encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "linenum",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 短路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	})

	// 实现两个判断日志等级的interface
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.InfoLevel
	})

	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	// 获取 info、error日志文件的io.Writer 抽象 getWriter() 在下方实现
	infoWriter := getWriter(filepath.Join(rootPath, "logs", "info.log"))
	errorWriter := getWriter(filepath.Join(rootPath, "logs", "error.log"))

	Alevel = zap.NewAtomicLevel()
	Alevel.SetLevel(zap.DebugLevel)
	// 最后创建具体的Logger
	var core zapcore.Core
	if rootPath != "no" {
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, zapcore.AddSync(infoWriter), infoLevel),
			zapcore.NewCore(encoder, zapcore.AddSync(errorWriter), errorLevel),
			//控制台输出日志
			zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), Alevel),
		)
	} else {
		core = zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), Alevel)
	}

	// 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 配置AddCallerSkip才能打印的原始行号
	log := zap.New(core, zap.AddCaller(), zap.Development(), zap.Fields(zap.String("serviceName", serviceName)), zap.AddCallerSkip(1))
	errorLogger = log.Sugar()

	resultLogger = zap.New(core, zap.AddCaller(), zap.Development(), zap.Fields(zap.String("serviceName", serviceName)), zap.AddCallerSkip(2)).Sugar()
}

func getWriter(filename string) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// demo.log是指向最新日志的链接
	// 保存7天内的日志，每1小时(整点)分割一次日志
	hook, err := rotatelogs.New(
		strings.Replace(filename, ".log", "", -1) + "-%Y%m%d%H.log", // 没有使用go风格反人类的format格式
		//rotatelogs.WithLinkName(filename),
		//rotatelogs.WithMaxAge(time.Hour*24*7),
		//rotatelogs.WithRotationTime(time.Hour),
	)
	if err != nil {
		panic(err)
	}

	return hook
}
func Debug(args ...interface{}) {
	errorLogger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	errorLogger.Debugf(template, args...)
}

func Info(args ...interface{}) {
	errorLogger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	errorLogger.Infof(template, args...)
}

func Warn(args ...interface{}) error {
	//go notice.SendDingMsg("Warn",args)
	errorLogger.Warn(args...)
	return fmt.Errorf("%v", args)
}

func Warnf(template string, args ...interface{}) {
	errorLogger.Warnf(template, args...)
}

func Error(args ...interface{}) error {
	if len(args) > 0 && args[0] == nil {
		return nil
	}
	errorLogger.Error(args...)
	return fmt.Errorf("%v", args)
}

func Errorf(template string, args ...interface{}) {
	errorLogger.Errorf(template, args...)
}

func DPanic(args ...interface{}) {
	errorLogger.DPanic(args...)
}

func DPanicf(template string, args ...interface{}) {
	errorLogger.DPanicf(template, args...)
}

func Panic(args ...interface{}) {
	errorLogger.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	errorLogger.Panicf(template, args...)
}

func Fatal(args ...interface{}) {
	errorLogger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	errorLogger.Fatalf(template, args...)
}

func ResultError(args ...interface{}) error {
	resultLogger.Error(args...)
	return fmt.Errorf("%v", args)
}
