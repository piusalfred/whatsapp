package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DecodeOptions struct {
	DisallowUnknownFields bool
	DisallowEmptyResponse bool
	InspectResponseError  bool
}

func DecodeResponseJSON[T any](response *http.Response, v *T, opts DecodeOptions) error {
	if response == nil {
		return ErrNilResponse
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	defer func() {
		response.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	}()

	if len(responseBody) == 0 && opts.DisallowEmptyResponse {
		return fmt.Errorf("%w: expected non-empty response body", ErrEmptyResponseBody)
	}

	isResponseOk := response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusPermanentRedirect

	if !isResponseOk { //nolint:nestif // no need to nest
		if opts.InspectResponseError {
			if len(responseBody) == 0 {
				return fmt.Errorf("%w: status code: %d", ErrRequestFailure, response.StatusCode)
			}

			var errorResponse ResponseError
			if err = json.Unmarshal(responseBody, &errorResponse); err != nil {
				return fmt.Errorf("%w: %w, status code: %d", ErrDecodeErrorResponse, err, response.StatusCode)
			}

			return &errorResponse
		}

		return fmt.Errorf("%w: status code: %d", ErrRequestFailure, response.StatusCode)
	}

	//	At this point, we know the response is 2xx
	//	If the body is empty here, that means empty bodies are allowed, so we return early.
	//	If the body is not empty but `v` is nil, we return an error as there is no target to decode into.
	//	Otherwise, we proceed with decoding the body into `v` if the body is not empty and `v` is provided.

	if len(responseBody) == 0 {
		return nil
	}

	if v == nil {
		return ErrNilTarget
	}

	decoder := json.NewDecoder(bytes.NewReader(responseBody))
	if opts.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if err = decoder.Decode(v); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodeResponseBody, err)
	}

	return nil
}

func DecodeRequestJSON[T any](request *http.Request, v *T, opts DecodeOptions) error {
	if request == nil {
		return ErrNilResponse
	}

	requestBody, err := io.ReadAll(request.Body)
	if err != nil {
		return fmt.Errorf("read request: %w", err)
	}
	defer func() {
		request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}()

	if len(requestBody) == 0 && opts.DisallowEmptyResponse {
		return fmt.Errorf("%w: expected non-empty request body", ErrEmptyResponseBody)
	}

	//	At this point, we know:
	//	- If the body is empty, it's allowed based on `DisallowEmptyResponse`.
	//	- If the body is not empty and `v == nil`, return an error as there’s no target to decode into.
	//	- Otherwise, proceed with decoding into `v` if the body is not empty and `v` is provided.

	if len(requestBody) == 0 {
		return nil
	}

	if v == nil {
		return ErrNilTarget
	}

	decoder := json.NewDecoder(bytes.NewReader(requestBody))
	if opts.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if decodeErr := decoder.Decode(v); decodeErr != nil {
		return fmt.Errorf("%w: %w", ErrDecodeResponseBody, decodeErr)
	}

	return nil
}

type (
	ResponseDecoderFunc func(ctx context.Context, response *http.Response) error

	ResponseDecoder interface {
		Decode(ctx context.Context, response *http.Response) error
	}

	ResponseBodyReaderFunc func(ctx context.Context, reader io.Reader) error
)

func (decoder ResponseDecoderFunc) Decode(ctx context.Context, response *http.Response) error {
	return decoder(ctx, response)
}

func ResponseDecoderJSON[T any](v *T, options DecodeOptions) ResponseDecoderFunc {
	fn := ResponseDecoderFunc(func(_ context.Context, response *http.Response) error {
		if err := DecodeResponseJSON(response, v, options); err != nil {
			return fmt.Errorf("decode json: %w", err)
		}

		return nil
	})

	return fn
}

func BodyReaderResponseDecoder(fn ResponseBodyReaderFunc) ResponseDecoderFunc {
	return func(ctx context.Context, response *http.Response) error {
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		defer func() {
			response.Body = io.NopCloser(bytes.NewBuffer(responseBody))
		}()

		if err = fn(ctx, bytes.NewReader(responseBody)); err != nil {
			return err
		}

		return nil
	}
}
