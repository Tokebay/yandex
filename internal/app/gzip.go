package app

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что клиент поддерживает gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") {
			w.Header().Set("Content-Encoding", "gzip")

			// Создаем новый Gzip Writer для записи сжатых данных
			gz := gzip.NewWriter(w)
			defer gz.Close()

			// Создаем новый ResponseWriter, который будет записывать данные через Gzip Writer
			gzw := gzipResponseWriter{
				Writer:         gz,
				ResponseWriter: w,
			}
			// Вызываем next обработчик с новым ResponseWriter
			next.ServeHTTP(gzw, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
