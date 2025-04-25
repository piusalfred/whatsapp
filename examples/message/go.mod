module examples/message

go 1.24.2

require (
	github.com/piusalfred/whatsapp v0.0.30
	github.com/piusalfred/whatsapp/examples v0.0.0-20250425063411-27a880346edb
)

require github.com/joho/godotenv v1.5.1 // indirect

replace (
	github.com/piusalfred/whatsapp => ../../
	github.com/piusalfred/whatsapp/examples => ../
)
