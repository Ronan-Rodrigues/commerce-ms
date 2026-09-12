# ── Stage 1: Build ──────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Instala certificados SSL e git
RUN apk add --no-cache ca-certificates git

# Cache de dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia o código-fonte
COPY . .

# Compila binário estático e enxuto
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/auth-service ./cmd/server

# ── Stage 2: Runtime ────────────────────────────────────────
FROM alpine:3.19

WORKDIR /app

# Cria usuário não-root por segurança
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Instala certificados de autoridade raiz para conexões HTTPS
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/auth-service /app/auth-service

# Ajusta permissões
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8081

ENTRYPOINT ["/app/auth-service"]

