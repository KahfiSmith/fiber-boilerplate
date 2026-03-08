FROM golang:1.25.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /fiber-boilerplate ./cmd/api

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /fiber-boilerplate /usr/local/bin/fiber-boilerplate

EXPOSE 3000

CMD ["fiber-boilerplate"]
