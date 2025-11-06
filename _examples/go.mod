module github.com/piusalfred/whatsapp/examples

go 1.25.1

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/joho/godotenv v1.5.1
	github.com/modelcontextprotocol/go-sdk v1.1.0
	github.com/piusalfred/whatsapp v1.0.0
	github.com/piusalfred/whatsapp/extras/mcp v0.0.0-20250926133252-af4372717978
)

require (
	github.com/google/jsonschema-go v0.3.0 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
)

replace github.com/piusalfred/whatsapp => ../

replace github.com/piusalfred/whatsapp/extras/mcp => ../extras/mcp
