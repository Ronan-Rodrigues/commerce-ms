# commerce-ms

Plataforma de comércio construída como microsserviços em Go.

## Stack

| Camada | Tecnologia |
|---|---|
| Backend | Go 1.22+, Chi router |
| Autenticação | JWT RS256, Argon2id, Redis blacklist |
| Mensageria | Redis Pub/Sub |
| Tempo real | SSE (Server-Sent Events) |
| Banco de dados | PostgreSQL 16 + sqlc |
| Cache / Rate limit | Redis 7 |
| Gateway | Go (Chi) + reverse proxy |
| Load Balancer | nginx |
| Frontend | React + TypeScript + Vite + TanStack + Tailwind |
| Automação | n8n + WAHA (WhatsApp) |
| Containers | Docker + docker-compose |
| IaC | Terraform (AWS) |
| CI/CD | GitHub Actions + Railway |

## Arquitetura

```
Internet → nginx (LB) → API Gateway → serviços
                                     ├── auth-service
                                     ├── catalog-service
                                     ├── order-service
                                     ├── payment-service
                                     └── notify-service (SSE + n8n)
```

Documentação completa: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)

## Como executar localmente

### Pré-requisitos

- Go 1.22+
- Docker e Docker Compose
- Make

### 1. Configurar variáveis de ambiente

```bash
cp .env.example .env
# edite o .env com suas configurações
```

### 2. Subir a infraestrutura

```bash
make up
```

### 3. Rodar as migrations

```bash
make migrate
```

### 4. Iniciar os serviços

```bash
# Em terminais separados (ou usar um process manager)
cd services/auth-service    && go run ./cmd/server
cd services/catalog-service && go run ./cmd/server
cd services/order-service   && go run ./cmd/server
cd services/payment-service && go run ./cmd/server
cd services/notify-service  && go run ./cmd/server
cd gateway                  && go run ./cmd/server
```

### 5. Acessar

- API: `http://localhost` (via nginx → gateway)
- Redis: `make redis-cli`

## Testes

```bash
make test
```

## Lint

```bash
make lint
```

## Documentação

- [Arquitetura](docs/ARCHITECTURE.md)
- [API](docs/API.md)
- [Terraform](infrastructure/terraform/README.md)

## Licença

MIT

