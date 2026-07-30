package middleware

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("underlying ResponseWriter does not support hijacking")
	}

	return hijacker.Hijack()
}

func (r *statusRecorder) Flush() {
    if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
        flusher.Flush()
    }
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		id, _ := r.Context().Value(requestIDKey).(string)

		log.Printf(
			"[%s] %s %s %d %s",
			id,
			r.Method,
			r.URL.Path,
			recorder.status,
			time.Since(start),
		)
	})
}
