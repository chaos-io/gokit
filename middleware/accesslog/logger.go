package accesslog

import (
	"context"
	"net/http"

	"github.com/chaos-io/chaos/logs"
)

type Level uint8

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type LogFunc func(context.Context, Level, string, ...any)

type RouteResolver func(*http.Request) string

type Option func(*options)

type options struct {
	log          LogFunc
	resolveRoute RouteResolver
}

func WithLogFunc(log LogFunc) Option {
	return func(opts *options) {
		if log != nil {
			opts.log = log
		}
	}
}

func WithRouteResolver(resolve RouteResolver) Option {
	return func(opts *options) {
		opts.resolveRoute = resolve
	}
}

func buildOptions(opts []Option) options {
	result := options{log: defaultLog}
	for _, option := range opts {
		if option != nil {
			option(&result)
		}
	}
	return result
}

func defaultLog(_ context.Context, level Level, message string, fields ...any) {
	switch level {
	case LevelDebug:
		logs.Debugw(message, fields...)
	case LevelWarn:
		logs.Warnw(message, fields...)
	case LevelError:
		logs.Errorw(message, fields...)
	default:
		logs.Infow(message, fields...)
	}
}
