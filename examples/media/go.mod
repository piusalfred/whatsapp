module examples/media

go 1.24.2

require (
	github.com/piusalfred/whatsapp v0.0.34
	github.com/piusalfred/whatsapp/examples v0.0.0-20250703095559-34431e3ec7b0
)

require github.com/joho/godotenv v1.5.1 // indirect

replace github.com/piusalfred/whatsapp => ../../

replace github.com/piusalfred/whatsapp/examples => ../
