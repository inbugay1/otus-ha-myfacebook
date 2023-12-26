FROM golang:1.21 AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -o ./bin/app ./cmd/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/bin ./bin
COPY --from=builder /app/storage ./storage

CMD ["./bin/app"]
EXPOSE 9092/tcp