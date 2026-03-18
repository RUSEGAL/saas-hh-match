FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o tg-bot ./cmd/main.go

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/tg-bot .

ENV BOT_TOKEN=
ENV API_URL=http://api:8080
ENV API_TOKEN=
ENV DATABASE_URL=postgres://postgres:postgres@postgres:5432/bot?sslmode=disable
ENV REDIS_ADDR=redis:6379
ENV SCHEDULER_ENABLED=true
ENV MODE=polling

CMD ["./tg-bot"]
