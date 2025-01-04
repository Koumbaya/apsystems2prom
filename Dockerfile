FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/

WORKDIR /app/cmd
RUN go build -o /app/app

FROM alpine:3.18

WORKDIR /root/

COPY --from=builder /app/app .

CMD ["sh", "-c", "./app --port ${PORT}"]
