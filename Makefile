build:
	go build -v -race ./...

test:
	go test -v ./...

build-cli:
	go build -o bin/whatsapp cmd/main.go