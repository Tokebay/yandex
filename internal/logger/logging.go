package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log *zap.Logger

func Initialize(level string) error {
	// Преобразование строки уровня логирования в объект zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	//Создание конфигурации логгера в режиме "production".
	cfg := zap.NewProductionConfig()
	//Установка уровня логирования для конфигурации.
	cfg.Level = lvl
	// Построение логгера на основе конфигурации.
	zl, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	// Присваивание построенного логгера переменной Log.
	Log = zl

	return nil
}

// логирования запросов и ответов на уровне HTTP-обработчиков
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

func (rl *responseLogger) WriteHeader(code int) {
	rl.statusCode = code
	rl.ResponseWriter.WriteHeader(code)
}

func (rl *responseLogger) Write(data []byte) (int, error) {
	rl.contentLength = len(data)
	return rl.ResponseWriter.Write(data)
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
