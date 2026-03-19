FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ai-service ./cmd/main.go

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/ai-service .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/prompts ./prompts

ENV NATS_URL=nats://nats:4222
ENV API_WEBHOOK_URL=http://api:8080/ai/webhook/analyze
ENV API_WEBHOOK_MATCHES_URL=http://api:8080/ai/webhook/matches
ENV AI_API_KEY=
ENV AI_MODEL=deepseek-chat
ENV AI_BASE_URL=https://api.deepseek.com
ENV RESUME_WORKERS=0
ENV VACANCY_WORKERS=2
ENV MAIN_API_URL=http://api:8080
ENV HH_API_URL=https://api.hh.ru
ENV MATCH_SCORE_THRESHOLD=0.70
ENV MATCH_BATCH_SIZE=20

CMD ["./ai-service"]
