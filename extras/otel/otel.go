package otel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	whttp "github.com/piusalfred/whatsapp/pkg/http"
)

const (
	WhatsappRequestTypeKey     = attribute.Key("whatsapp.request.type")
	WhatsappRequestMethodKey   = attribute.Key("whatsapp.request.method")
	WhatsappRequestURLKey      = attribute.Key("whatsapp.request.url")
	WhatsappRequestIsSecureKey = attribute.Key("whatsapp.request.is_secure")
	WhatsappRequestErrorKey    = attribute.Key("whatsapp.request.error")
)

const tracerName = "github.com/piusalfred/whatsapp/extras/otel"

type Options struct {
	traceProvider  trace.TracerProvider
	meterProvider  metric.MeterProvider
	textPropagator propagation.TextMapPropagator
	resource       *resource.Resource
	attributes     []attribute.KeyValue
}

type Option func(*Options)

func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(o *Options) {
		o.traceProvider = tp
	}
}

func WithMeter(meter metric.MeterProvider) Option {
	return func(o *Options) {
		o.meterProvider = meter
	}
}

func WithPropagator(p propagation.TextMapPropagator) Option {
	return func(o *Options) {
		o.textPropagator = p
	}
}

func WithResource(r *resource.Resource) Option {
	return func(o *Options) {
		o.resource = r
	}
}

func WithSpanAttributes(attrs ...attribute.KeyValue) Option {
	return func(o *Options) {
		o.attributes = attrs
	}
}

type Sender[T any] struct {
	core            whttp.Sender[T]
	tracer          trace.Tracer
	requestsCounter metric.Int64Counter
	errorsCounter   metric.Int64Counter
}

func NewSender[T any](core whttp.Sender[T], opts ...Option) (*Sender[T], error) {
	options := Options{
		traceProvider:  tracenoop.NewTracerProvider(),
		meterProvider:  metricnoop.NewMeterProvider(),
		textPropagator: propagation.TraceContext{},
		resource:       resource.Empty(),
		attributes:     []attribute.KeyValue{},
	}

	for _, opt := range opts {
		opt(&options)
	}

	tracer := options.traceProvider.Tracer(
		tracerName,
		trace.WithInstrumentationAttributes(options.attributes...),
	)

	meter := options.meterProvider.Meter(
		tracerName,
		metric.WithInstrumentationAttributes(options.attributes...),
	)

	requestCounter, err := meter.Int64Counter(
		"requests_total",
		metric.WithDescription("Total number of requests sent"),
	)
	if err != nil {
		return nil, err
	}

	errorCounter, err := meter.Int64Counter(
		"errors_total",
		metric.WithDescription("Total number of errors encountered"),
	)
	if err != nil {
		return nil, err
	}

	return &Sender[T]{
		core:            core,
		tracer:          tracer,
		requestsCounter: requestCounter,
		errorsCounter:   errorCounter,
	}, nil
}

func (s *Sender[T]) Send(ctx context.Context, req *whttp.Request[T], decoder whttp.ResponseDecoder) error {
	attrs := []attribute.KeyValue{
		WhatsappRequestTypeKey.String(req.Type.String()),
		WhatsappRequestMethodKey.String(req.Method),
		WhatsappRequestIsSecureKey.Bool(req.SecureRequests),
	}

	reqURL, err := req.URL()
	if err == nil {
		WhatsappRequestURLKey.String(reqURL.String())
	}

	ctx, span := s.tracer.Start(ctx, req.Method,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)

	defer span.End()

	s.requestsCounter.Add(ctx, 1, metric.WithAttributes(attrs...))

	if err = s.core.Send(ctx, req, decoder); err != nil {
		attrs = append(attrs, WhatsappRequestErrorKey.String(err.Error()))
		span.RecordError(err)
		s.errorsCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	span.SetStatus(codes.Ok, "OK")

	return nil
}
