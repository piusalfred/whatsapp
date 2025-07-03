module github.com/piusalfred/whatsapp/extras/otel

go 1.24.0

replace github.com/piusalfred/whatsapp v1.0.0 => ../..

require (
	github.com/piusalfred/whatsapp v1.0.0
	go.opentelemetry.io/otel v1.37.0
	go.opentelemetry.io/otel/metric v1.37.0
	go.opentelemetry.io/otel/trace v1.37.0
)
