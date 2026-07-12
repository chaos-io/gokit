package accesslog

import (
	"io"
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
)

func HTTPMiddleware(cfg Config, options ...Option) func(http.Handler) http.Handler {
	policy := newPolicy(cfg)
	opts := buildOptions(options)

	return func(next http.Handler) http.Handler {
		if next == nil {
			next = http.NotFoundHandler()
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			statusCode := http.StatusOK
			var bytesWritten int64
			wroteHeader := false

			wrapped := httpsnoop.Wrap(w, httpsnoop.Hooks{
				WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
					return func(code int) {
						if code < 100 || code > 199 {
							if !wroteHeader {
								statusCode = code
								wroteHeader = true
							}
						}
						next(code)
					}
				},
				Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
					return func(body []byte) (int, error) {
						n, err := next(body)
						bytesWritten += int64(n)
						wroteHeader = true
						return n, err
					}
				},
				ReadFrom: func(next httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
					return func(src io.Reader) (int64, error) {
						n, err := next(src)
						bytesWritten += n
						wroteHeader = true
						return n, err
					}
				},
			})

			defer func() {
				duration := time.Since(started)
				recovered := recover()
				if recovered != nil {
					statusCode = http.StatusInternalServerError
				}
				logHTTP(r, statusCode, bytesWritten, duration, policy, opts)
				if recovered != nil {
					panic(recovered)
				}
			}()

			next.ServeHTTP(wrapped, r)
		})
	}
}

func logHTTP(r *http.Request, statusCode int, bytesWritten int64, duration time.Duration, policy *policy, opts options) {
	path := ""
	method := ""
	if r != nil {
		method = r.Method
		if r.URL != nil {
			path = r.URL.Path
		}
	}
	requestID := httpRequestID(r)
	if !policy.ShouldLog(Event{
		Protocol:  ProtocolHTTP,
		Operation: path,
		RequestID: requestID,
		Duration:  duration,
		Important: importantHTTPStatus(statusCode),
	}) {
		return
	}

	fields := []any{
		"protocol", ProtocolHTTP,
		"method", method,
		"path", path,
		"status", statusCode,
		"bytes", bytesWritten,
		"duration", duration.String(),
		"duration_ms", float64(duration.Microseconds()) / 1000,
		"remote_ip", httpRemoteIP(r),
		"request_id", requestID,
	}
	if opts.resolveRoute != nil {
		if route := opts.resolveRoute(r); route != "" {
			fields = append(fields, "route", route)
		}
	}
	if r != nil {
		fields = append(fields, traceFields(r.Context())...)
	}
	opts.log(r.Context(), httpLevel(statusCode), "http access", fields...)
}

func importantHTTPStatus(code int) bool {
	if code >= http.StatusInternalServerError {
		return true
	}
	switch code {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests:
		return true
	default:
		return false
	}
}

func httpLevel(code int) Level {
	switch {
	case code >= http.StatusInternalServerError:
		return LevelError
	case code >= http.StatusBadRequest:
		return LevelWarn
	default:
		return LevelInfo
	}
}
