package config

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	Level      string // debug, info, warn, prod
	Format     string // local(見やすさ重視), cloud(CloudWatch等で解析可能であることを重視)
	AppName    string // アプリ名(cloudでのみログ出力)
	AppVersion string // アプリのバージョン(cloudでのみログ出力)
}

type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
}

type zapLogger struct {
	l *zap.Logger
}

func (z *zapLogger) Debug(msg string, fields ...zap.Field) {
	z.l.Debug(msg, fields...)
}
func (z *zapLogger) Info(msg string, fields ...zap.Field) {
	z.l.Info(msg, fields...)
}
func (z *zapLogger) Warn(msg string, fields ...zap.Field) {
	z.l.Warn(msg, fields...)
}
func (z *zapLogger) Error(msg string, fields ...zap.Field) {
	z.l.Error(msg, fields...)
}

// NewLogger creates a new Logger based on LoggerConfig
func NewLogger(cfg LoggerConfig) (Logger, error) {
	var zapCfg zap.Config
	switch cfg.Format {
	case "local":
		// local環境では読みやすさ重視
		// (非構造化ログ、JST、ミリ秒精度)
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			jst := t.In(time.FixedZone("Asia/Tokyo", 9*60*60))
			enc.AppendString(jst.Format("2006-01-02T15:04:05.000Z07:00"))
		}
	case "cloud":
		// cloud環境ではcloud watch等で読まれる前提で解析重視
		// (構造化ログ、UTC、ナノ秒精度)
		zapCfg = zap.NewProductionConfig()
		zapCfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339Nano)
	default:
		// LOG_FORMATが不正の場合、cloud向けフォーマットで出力
		fmt.Fprintf(os.Stderr, "invalid LOG_FORMAT %q, fallback to 'cloud'\n", cfg.Format)
		zapCfg = zap.NewProductionConfig()
		zapCfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339Nano)
	}

	// ログレベル設定
	level := cfg.Level
	parsedLevel := zapcore.InfoLevel
	if err := parsedLevel.UnmarshalText([]byte(level)); err != nil {
		fmt.Fprintf(os.Stderr, "invalid LOG_LEVEL %q, fallback to 'info'\n", level)
		parsedLevel = zapcore.InfoLevel
	}
	zapCfg.Level = zap.NewAtomicLevelAt(parsedLevel)

	// error時のみStackTrace出力するよう設定
	zapCfg.DisableStacktrace = true
	baseLogger, err := zapCfg.Build(
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to build logger: %w", err)
		fmt.Fprintln(os.Stderr, wrappedErr)
		return nil, err
	}

	// cloud向けはフィールド追加
	if cfg.Format == "cloud" {
		baseLogger = baseLogger.With(
			zap.String("app", cfg.AppName),
			zap.String("version", cfg.AppVersion),
		)
	}

	return &zapLogger{l: baseLogger}, nil
}
