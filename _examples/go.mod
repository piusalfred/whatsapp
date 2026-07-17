module github.com/piusalfred/whatsapp/examples

go 1.26.5

require (
	github.com/go-chi/chi/v5 v5.3.1
	github.com/joho/godotenv v1.5.1
	github.com/piusalfred/whatsapp v0.0.0
)

require golang.org/x/sync v0.22.0 // indirect

replace github.com/piusalfred/whatsapp v0.0.0 => ../
