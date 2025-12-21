package http

import (
	"fmt"
	"strings"

	werrors "github.com/piusalfred/whatsapp/pkg/errors"
)

type ResponseError struct {
	Code int            `json:"code,omitempty"`
	Err  *werrors.Error `json:"error,omitempty"`
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("whatsapp message error: http code: %d, %s", e.Code, strings.ToLower(e.Err.Error()))
}

func (e *ResponseError) Unwrap() error {
	return e.Err
}

const (
	ErrNilRequest          = httpError("nil request provided")
	ErrNilResponse         = httpError("nil response provided")
	ErrEmptyResponseBody   = httpError("empty response body")
	ErrNilTarget           = httpError("nil value passed for decoding target")
	ErrRequestFailure      = httpError("request failed")
	ErrDecodeResponseBody  = httpError("failed to decode response body")
	ErrDecodeErrorResponse = httpError("failed to decode error response")
)

type httpError string

func (e httpError) Error() string {
	return string(e)
}
