package cradle

import "go.uber.org/zap"

type GoLetToZapLogger struct{}

// Implements io.Writer
func (_ GoLetToZapLogger) Write(p []byte) (n int, err error) {
	zap.L().Info("go-let", zap.String("msg", string(p)))
	return len(p), nil
}
