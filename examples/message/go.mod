module examples/message

go 1.24.2

require (
	github.com/piusalfred/whatsapp v0.0.33
	github.com/piusalfred/whatsapp/examples v0.0.0-20250703092454-b85fe47ee7b6
)

require github.com/joho/godotenv v1.5.1 // indirect

replace (
	github.com/piusalfred/whatsapp => ../../
	github.com/piusalfred/whatsapp/examples => ../
)
