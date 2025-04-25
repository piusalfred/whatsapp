module examples/message

go 1.24.2

require (
	github.com/piusalfred/whatsapp v0.0.29
	github.com/piusalfred/whatsapp/examples v0.0.0-00010101000000-000000000000
)

require github.com/joho/godotenv v1.5.1 // indirect

replace (
	github.com/piusalfred/whatsapp => ../../
	github.com/piusalfred/whatsapp/examples => ../
)
