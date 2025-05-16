package config

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	Level      string         // debug, info, warn, error
	Format     string         // local(見やすさ重視), cloud(CloudWatch等で解析可能であることを重視)
	AppName    string         // アプリ名(cloudでのみログ出力)
	AppVersion string         // アプリのバージョン(cloudでのみログ出力)
	Location   *time.Location // ログ出力時のタイムゾーン(nilならUTC)
}

func NewLogger(cfg LoggerConfig) (*zap.Logger, error) {
	var zapCfg zap.Config
	switch cfg.Format {
	case "local":
		// local環境では読みやすさ重視
		// (非構造化ログ、JST、ミリ秒精度)
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			loc := cfg.Location
			if loc == nil {
				loc = time.UTC
			}
			enc.AppendString(t.In(loc).Format("2006-01-02T15:04:05.000Z07:00"))
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
	parsedLevel := zapcore.InfoLevel
	if err := parsedLevel.UnmarshalText([]byte(cfg.Level)); err != nil {
		fmt.Fprintf(os.Stderr, "invalid LOG_LEVEL %q, fallback to 'info'\n", cfg.Level)
	}
	zapCfg.Level = zap.NewAtomicLevelAt(parsedLevel)

	// error時のみStackTrace出力するよう設定
	zapCfg.DisableStacktrace = true
	logger, err := zapCfg.Build(
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	// cloud向けはフィールド追加
	if cfg.Format == "cloud" {
		logger = logger.With(
			zap.String("app", cfg.AppName),
			zap.String("version", cfg.AppVersion),
		)
	}

	return logger, nil
}
