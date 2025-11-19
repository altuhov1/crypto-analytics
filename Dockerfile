FROM golang:1.25-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN mkdir -p bin

RUN go build -o bin/httpBack ./cmd/httpBack

COPY static ./static
COPY storage ./storage

EXPOSE 8080

CMD ["./bin/httpBack"]
