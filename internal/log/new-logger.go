package log

import (
	"github.com/natefinch/lumberjack"
	"github.com/samber/lo"
	"github.com/snivilised/extendio/xfs/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/snivilised/extendio/i18n"
)

func NewLogger(info *LoggerInfo) Ref {
	return utils.NewRoProp(lo.TernaryF(info.Enabled,
		func() Logger {
			if info.Path == "" {
				panic(i18n.NewInvalidConfigEntryError(info.Path, "-"))
			}
			ws := zapcore.AddSync(&lumberjack.Logger{
				Filename:   info.Path,
				MaxSize:    info.Rotation.MaxSizeInMb,
				MaxBackups: info.Rotation.MaxNoOfBackups,
				MaxAge:     info.Rotation.MaxAgeInDays,
			})
			config := zap.NewProductionEncoderConfig()
			config.EncodeTime = zapcore.TimeEncoderOfLayout(info.TimeStampFormat)
			core := zapcore.NewCore(
				zapcore.NewJSONEncoder(config),
				ws,
				info.Level,
			)
			return zap.New(core)
		}, func() Logger {
			return zap.NewNop()
		}),
	)
}
