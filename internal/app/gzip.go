package app

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip {
			w.Header().Set("Content-Encoding", "gzip")
			fmt.Printf("acceptEncoding %s\n", acceptEncoding)
			// Создаем Gzip Writer для записи сжатых данных
			gz := gzip.NewWriter(w)
			defer gz.Close()

			// Создаем новый ResponseWriter, который будет записывать данные через Gzip Writer
			gzw := gzipResponseWriter{
				Writer:         gz,
				ResponseWriter: w,
			}

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
