package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log *zap.Logger

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	// lvl.UnmarshalText([]byte(level))

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	Log = zl

	return nil
}
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		recorder := &responseLogger{ResponseWriter: w}
		next.ServeHTTP(recorder, r)

		duration := time.Since(startTime)
		Log.Info("Request handled",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", duration),
			zap.Int("status_code", recorder.statusCode),
			zap.Int("content_length", recorder.contentLength),
		)
	})
}

type responseLogger struct {
	http.ResponseWriter
	statusCode    int
	contentLength int
}

func (l *responseLogger) WriteHeader(code int) {
	l.statusCode = code
	l.ResponseWriter.WriteHeader(code)
}

func (l *responseLogger) Write(data []byte) (int, error) {
	l.contentLength = len(data)
	return l.ResponseWriter.Write(data)
}
