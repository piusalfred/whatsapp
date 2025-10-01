package otel

import (
	"context"
	"iter"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	whttp "github.com/piusalfred/whatsapp/pkg/http"
	"github.com/piusalfred/whatsapp/webhooks"
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
	traceProvider trace.TracerProvider
	meterProvider metric.MeterProvider
	attributes    []attribute.KeyValue
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
		traceProvider: tracenoop.NewTracerProvider(),
		meterProvider: metricnoop.NewMeterProvider(),
		attributes:    []attribute.KeyValue{},
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
		attrs = append(attrs, WhatsappRequestURLKey.String(reqURL))
	}

	ctx, span := s.tracer.Start(ctx, req.Type.Name(),
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

var _ webhooks.NotificationHandler = (*WebhookHandler)(nil)

type WebhookHandler struct {
	tracer trace.Tracer
	next   webhooks.NotificationHandler
}

func NewWebhookHandler(tracer trace.Tracer, next webhooks.NotificationHandler) *WebhookHandler {
	return &WebhookHandler{
		tracer: tracer,
		next:   next,
	}
}

func NotificationLogValues(notification *webhooks.Notification) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("webhook.notification.object", notification.Object),
	}

	var entryIDs []string
	var changeFields []string

	for _, entry := range notification.Entry {
		entryIDs = append(entryIDs, entry.ID)
		for _, change := range entry.Changes {
			changeFields = append(changeFields, change.Field)
		}
	}

	if len(entryIDs) > 0 {
		attrs = append(attrs, normalizeAttr("webhook.notification.entries", entryIDs))
	}

	if len(changeFields) > 0 {
		attrs = append(attrs, normalizeAttr("webhook.notification.changes", changeFields))
	}

	return attrs
}

func normalizeAttr(key string, values []string) attribute.KeyValue {
	if len(values) == 1 {
		return attribute.String(key, values[0])
	}

	return attribute.StringSlice(key, values)
}

const (
	WebhookMessageIDKey        = attribute.Key("webhook.notification.message.id")
	WebhookMessageTypeKey      = attribute.Key("webhook.notification.message.type")
	WebhookMessageStatusKey    = attribute.Key("webhook.notification.message.status")
	WebhookMessageTimestampKey = attribute.Key("webhook.notification.message.timestamp")
	WebhookMessageReplyKey     = attribute.Key("webhook.notification.message.reply")
	WebhookMessageSenderKey    = attribute.Key("webhook.notification.message.sender")
	WebhookMessageForwardedKey = attribute.Key("webhook.notification.message.forwarded")
)

func (o *WebhookHandler) HandleNotification(
	ctx context.Context,
	notification *webhooks.Notification,
) *webhooks.Response {
	ctx, span := o.tracer.Start(ctx, "HandleNotification",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			NotificationLogValues(notification)...,
		))

	defer span.End()

	for status := range statusValueIter(notification) {
		span.AddEvent("message status updated",
			trace.WithAttributes(
				WebhookMessageIDKey.String(status.ID),
				WebhookMessageStatusKey.String(status.StatusValue),
				WebhookMessageTimestampKey.String(status.Timestamp),
			),
		)
	}

	for message := range messageValueIter(notification) {
		span.AddEvent("message received",
			trace.WithAttributes(
				WebhookMessageIDKey.String(message.ID),
				WebhookMessageTypeKey.String(message.Type),
				WebhookMessageTimestampKey.String(message.Timestamp),
				WebhookMessageReplyKey.Bool(message.IsAReply()),
				WebhookMessageSenderKey.String(message.From),
				WebhookMessageForwardedKey.Bool(message.IsForwarded()),
			),
		)
	}

	response := o.next.HandleNotification(ctx, notification)

	if response.StatusCode != http.StatusOK {
		span.SetStatus(codes.Error, "error handling notification")
	} else {
		span.SetStatus(codes.Ok, "OK")
	}

	span.SetAttributes(attribute.Int("webhook.notification.response.status", response.StatusCode))

	return response
}

func statusValueIter(n *webhooks.Notification) iter.Seq[webhooks.Status] {
	return func(yield func(webhooks.Status) bool) {
		statuses := make([]webhooks.Status, 0)
		for _, entry := range n.Entry {
			for _, change := range entry.Changes {
				if change.Value != nil {
					statusValues := PtrToValueList(change.Value.Statuses)
					statuses = append(statuses, statusValues...)
				}
			}
		}

		for _, status := range statuses {
			if !yield(status) {
				return
			}
		}
	}
}

func messageValueIter(n *webhooks.Notification) iter.Seq[webhooks.Message] {
	return func(yield func(webhooks.Message) bool) {
		messages := make([]webhooks.Message, 0)
		for _, entry := range n.Entry {
			for _, change := range entry.Changes {
				if change.Value != nil {
					messageValues := PtrToValueList(change.Value.Messages)
					messages = append(messages, messageValues...)
				}
			}
		}

		for _, message := range messages {
			if !yield(message) {
				return
			}
		}
	}
}

func PtrToValue[T any](ptr *T) T {
	if ptr == nil {
		var zeroValue T
		return zeroValue
	}
	return *ptr
}

func PtrToValueList[T any](ptr []*T) []T {
	if ptr == nil {
		return nil
	}
	values := make([]T, len(ptr))
	for i, v := range ptr {
		values[i] = PtrToValue(v)
	}
	return values
}
