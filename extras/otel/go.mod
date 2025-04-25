module github.com/piusalfred/whatsapp/extras/otel

go 1.24.0

replace github.com/piusalfred/whatsapp v0.0.29 => ../..

require (
	github.com/piusalfred/whatsapp v0.0.29
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/metric v1.35.0
	go.opentelemetry.io/otel/trace v1.35.0
)
