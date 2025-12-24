package logger

import (
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 全局logger实例
// 默认为 Nop，避免在未调用 Init 的场景（如部分测试）发生空指针 panic。
var Logger *zap.Logger = zap.NewNop()

// Init 初始化logger
// logDir: 日志文件所在目录（通常是可执行文件所在目录）
// logLevel: 日志级别（debug, info, warn, error）
func Init(logDir string, logLevel string) error {
	// 创建log子目录
	logSubDir := filepath.Join(logDir, "log")
	if err := os.MkdirAll(logSubDir, 0755); err != nil {
		return err
	}

	logFile := filepath.Join(logSubDir, "app.log")

	// lumberjack 日志轮转配置
	writer := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10,   // MB
		MaxBackups: 5,    // 保留5个备份
		MaxAge:     30,   // 保留30天
		Compress:   true, // 压缩旧日志
	}

	// 同时输出到文件和控制台
	multiWriter := zapcore.AddSync(io.MultiWriter(writer, os.Stdout))

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
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
	}

	// 解析日志级别，默认为info
	level := zapcore.InfoLevel
	if logLevel != "" {
		if parsedLevel, err := zapcore.ParseLevel(logLevel); err == nil {
			level = parsedLevel
		}
	}

	// 创建core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		multiWriter,
		level,
	)

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}
