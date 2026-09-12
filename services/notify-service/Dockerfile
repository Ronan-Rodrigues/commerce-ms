# ── Stage 1: Build ──────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/notify-service ./cmd/server

# ── Stage 2: Runtime ────────────────────────────────────────
FROM alpine:3.19

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/notify-service /app/notify-service

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8085

ENTRYPOINT ["/app/notify-service"]

