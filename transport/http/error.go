package http

import (
	"errors"
	"net/http"

	"github.com/chaos-io/core/go/chaos/core"
	pkgerrors "github.com/pkg/errors"
)

type codedError interface {
	error
	Code() int32
	Message() string
	Extra() map[string]string
}

// Error satisfies the Headerer and StatusCoder interfaces in
// package github.com/go-kit/kit/transport/http.
type Error struct {
	error
	statusCode int
	headers    http.Header
}

func WrapError(e error, code int, message string, headers ...string) *Error {
	err := &Error{
		error:      pkgerrors.Wrap(e, message),
		statusCode: code,
		headers:    make(http.Header),
	}

	length := len(headers)
	if length > 0 && length%2 == 0 {
		for i := 0; i < length; i += 2 {
			err.headers.Add(headers[i], headers[i+1])
		}
	}
	return err
}

func CoreErrorFromError(err error) *core.Error {
	if err == nil {
		return nil
	}

	var coreErr *core.Error
	if errors.As(err, &coreErr) {
		return coreErr
	}

	var coded codedError
	if errors.As(err, &coded) {
		return core.NewErrorFrom(coded.Code(), coded.Message())
	}

	return core.NewErrorFrom(http.StatusInternalServerError, err.Error())
}

func (e Error) StatusCode() int {
	return e.statusCode
}

func (e Error) Headers() http.Header {
	return e.headers
}
