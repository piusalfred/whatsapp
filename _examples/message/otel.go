/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the "Software"), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/piusalfred/whatsapp/message"
	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const instrumentationName = "github.com/piusalfred/whatsapp/_examples/message"

type Telemetry struct {
	Logger         *slog.Logger
	Propagator     propagation.TextMapPropagator
	TraceProvider  *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	tracer         trace.Tracer
	sendCounter    metric.Int64Counter
	sendDuration   metric.Float64Histogram
	sendErrorCount metric.Int64Counter
	shutdownFunc   func(context.Context) error
}

// InitTelemetry initializes the OTel pipeline along with its corresponding standard
// metrics, tracers, and loggers. It groups them cleanly inside a Telemetry manager instance.
func InitTelemetry(ctx context.Context) (*Telemetry, error) {
	telemetry := &Telemetry{
		Logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	var shutdownFuncs []func(context.Context) error

	shutdown := func(ctx context.Context) error {
		var shutdownErr error
		for i := len(shutdownFuncs) - 1; i >= 0; i-- {
			shutdownErr = errors.Join(shutdownErr, shutdownFuncs[i](ctx))
		}
		shutdownFuncs = nil
		return shutdownErr
	}

	handleErr := func(inErr error) error {
		return errors.Join(inErr, shutdown(context.Background()))
	}

	// Propagator setup: W3C TraceContext + Baggage.
	telemetry.Propagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(telemetry.Propagator)

	// Trace provider initialization
	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, handleErr(err)
	}
	telemetry.TraceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter, sdktrace.WithBatchTimeout(time.Second)),
	)
	shutdownFuncs = append(shutdownFuncs, telemetry.TraceProvider.Shutdown)
	otel.SetTracerProvider(telemetry.TraceProvider)

	// Meter provider initialization
	metricExporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		return nil, handleErr(err)
	}
	telemetry.MeterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(3*time.Second))),
	)
	shutdownFuncs = append(shutdownFuncs, telemetry.MeterProvider.Shutdown)
	otel.SetMeterProvider(telemetry.MeterProvider)

	// Instruments setup
	telemetry.tracer = otel.Tracer(instrumentationName)
	meter := otel.Meter(instrumentationName)

	telemetry.sendCounter, err = meter.Int64Counter(
		"whatsapp.message.sends",
		metric.WithDescription("Number of WhatsApp message send operations"),
		metric.WithUnit("{send}"),
	)
	if err != nil {
		return nil, handleErr(fmt.Errorf("create send counter: %w", err))
	}

	telemetry.sendDuration, err = meter.Float64Histogram(
		"whatsapp.message.send_duration",
		metric.WithDescription("Duration of WhatsApp message send operations"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, handleErr(fmt.Errorf("create send duration histogram: %w", err))
	}

	telemetry.sendErrorCount, err = meter.Int64Counter(
		"whatsapp.message.send_errors",
		metric.WithDescription("Number of failed WhatsApp message send operations"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return nil, handleErr(fmt.Errorf("create send error counter: %w", err))
	}

	telemetry.shutdownFunc = shutdown

	return telemetry, nil
}

// Close flushes the telemetry providers and stops metrics tracking safely.
func (t *Telemetry) Close(ctx context.Context) error {
	if t.shutdownFunc != nil {
		return t.shutdownFunc(ctx)
	}
	return nil
}

// Middleware unifies structured logs, OTel tracing spans, and runtime counters
// directly into a cohesive messaging middleware block.
func (t *Telemetry) Middleware() whttp.Middleware[message.BaseRequest] {
	return func(next whttp.SenderFunc[message.BaseRequest]) whttp.SenderFunc[message.BaseRequest] {
		return func(ctx context.Context, req *whttp.Request[message.BaseRequest], decoder whttp.ResponseDecoder) error {
			spanName := req.Type.String()
			if spanName == "" {
				spanName = "whatsapp.message.send"
			}

			ctx, span := t.tracer.Start(ctx, spanName)
			defer span.End()

			span.SetAttributes(
				semconv.HTTPRequestMethodKey.String(req.Method),
				attribute.String("whatsapp.request.type", req.Type.String()),
			)

			span.AddEvent("sending",
				trace.WithAttributes(
					attribute.String("whatsapp.endpoint", fmt.Sprintf("%v", req.Endpoints)),
				),
			)

			t.Logger.Info("sending message", "request", req)

			reqTypeAttr := attribute.String("whatsapp.request.type", req.Type.String())
			start := time.Now()

			err := next(ctx, req, decoder)

			duration := time.Since(start).Seconds()

			if t.sendDuration.Enabled(ctx) {
				t.sendDuration.Record(ctx, duration, metric.WithAttributes(reqTypeAttr))
			}

			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				span.AddEvent("send failed",
					trace.WithAttributes(
						attribute.String("error.message", err.Error()),
					),
				)

				if t.sendErrorCount.Enabled(ctx) {
					t.sendErrorCount.Add(ctx, 1, metric.WithAttributes(reqTypeAttr))
				}

				t.Logger.Error("send message", "error", err)
				return err
			}

			span.AddEvent("send completed")
			if t.sendCounter.Enabled(ctx) {
				t.sendCounter.Add(ctx, 1, metric.WithAttributes(reqTypeAttr))
			}

			t.Logger.Info("message sent successfully")
			return nil
		}
	}
}

func OTelHTTPTransport(conf *TransportParams) http.RoundTripper {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
	var opts []otelhttp.Option
	if conf != nil {
		if conf.Propagators != nil {
			opts = append(opts, otelhttp.WithPropagators(conf.Propagators))
		}
		if conf.MeterProvider != nil {
			opts = append(opts, otelhttp.WithMeterProvider(conf.MeterProvider))
		}
		if conf.TracerProvider != nil {
			opts = append(opts, otelhttp.WithTracerProvider(conf.TracerProvider))
		}
		if conf.ServerName != "" {
			opts = append(opts, otelhttp.WithServerName(conf.ServerName))
		}
	}
	return otelhttp.NewTransport(transport, opts...)
}

type TransportParams struct {
	Propagators    propagation.TextMapPropagator
	MeterProvider  *sdkmetric.MeterProvider
	TracerProvider *sdktrace.TracerProvider
	ServerName     string
}
