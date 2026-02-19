module github.com/piusalfred/whatsapp/examples

go 1.25.1

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/joho/godotenv v1.5.1
	github.com/modelcontextprotocol/go-sdk v1.3.1
	github.com/piusalfred/whatsapp v1.0.0
	github.com/piusalfred/whatsapp/extras/mcp v0.0.0-20251221095014-a550cfbe294b
)

require (
	github.com/google/jsonschema-go v0.4.2 // indirect
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/segmentio/encoding v0.5.3 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/oauth2 v0.34.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
)

replace github.com/piusalfred/whatsapp => ../

replace github.com/piusalfred/whatsapp/extras/mcp => ../extras/mcp
