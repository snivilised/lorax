package log

import (
	"go.uber.org/zap"
)

func String(key, val string) Field {
	return zap.String(key, val)
}

func Uint(key string, val uint) Field {
	return zap.Uint(key, val)
}

func Int(key string, val int) Field {
	return zap.Int(key, val)
}

func Float64(key string, val float64) Field {
	return zap.Float64(key, val)
}
