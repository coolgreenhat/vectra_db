FROM golang:1.24.7-alpine

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o vectordbd ./cmd/vectordbd 

EXPOSE 8080

CMD ["./vectordbd"]
