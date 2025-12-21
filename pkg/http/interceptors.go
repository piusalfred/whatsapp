package http

import (
	"context"
	"net/http"
)

type (
	RequestInterceptorFunc func(ctx context.Context, request *http.Request) error
	RequestInterceptor     interface {
		InterceptRequest(ctx context.Context, request *http.Request) error
	}

	ResponseInterceptorFunc func(ctx context.Context, response *http.Response) error
	ResponseInterceptor     interface {
		InterceptResponse(ctx context.Context, response *http.Response) error
	}
)

func (fn RequestInterceptorFunc) InterceptRequest(ctx context.Context, request *http.Request) error {
	return fn(ctx, request)
}

func (fn ResponseInterceptorFunc) InterceptResponse(ctx context.Context, response *http.Response) error {
	return fn(ctx, response)
}
