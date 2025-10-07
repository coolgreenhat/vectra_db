run:
	go run ./cmd/vectordbd

build:
	go build -o bin/vectordbd ./cmd/vectordbd 

test:
	go test ./... -v 
