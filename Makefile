# ─────────────────────────────────────────────────────────────
# commerce-ms — Makefile
# Atalhos para o dia a dia de desenvolvimento
# Uso: make <comando>
# ─────────────────────────────────────────────────────────────

.PHONY: help up down logs ps redis-cli test lint build migrate

# Mostra todos os comandos disponíveis
help:
	@echo ""
	@echo "  commerce-ms — Comandos disponíveis"
	@echo "  ─────────────────────────────────────────"
	@echo "  make up          Sobe PostgreSQL, Redis e nginx"
	@echo "  make down        Derruba todos os containers"
	@echo "  make logs        Exibe logs de todos os serviços"
	@echo "  make ps          Lista containers em execução"
	@echo "  make redis-cli   Abre o redis-cli no container"
	@echo "  make test        Roda testes de todos os serviços"
	@echo "  make lint        Executa golangci-lint em todos"
	@echo "  make build       Compila todos os binários Go"
	@echo "  make migrate     Roda migrations de todos os serviços"
	@echo ""

# ── Docker ──────────────────────────────────────────────────

# Sobe a infraestrutura base (banco, redis, nginx)
up:
	docker compose up -d
	@echo "✅ PostgreSQL, Redis e nginx prontos"

# Derruba tudo e remove orphans
down:
	docker compose down --remove-orphans

# Logs em tempo real
logs:
	docker compose logs -f

# Status dos containers
ps:
	docker compose ps

# Abre o redis-cli para inspecionar dados
redis-cli:
	docker compose exec redis redis-cli

# ── Go ───────────────────────────────────────────────────────

# Roda todos os testes com detecção de race conditions
test:
	@for service in auth-service catalog-service order-service payment-service notify-service; do \
		echo "🧪 Testando $$service..."; \
		cd services/$$service && go test ./... -race -count=1 -coverprofile=coverage.out && cd ../..; \
	done

# Lint em todos os serviços
lint:
	@for service in auth-service catalog-service order-service payment-service notify-service gateway; do \
		echo "🔍 Lint em $$service..."; \
		cd services/$$service && golangci-lint run ./... && cd ../..; \
	done

# Compila os binários de todos os serviços
build:
	@for service in auth-service catalog-service order-service payment-service notify-service; do \
		echo "🔨 Compilando $$service..."; \
		cd services/$$service && go build -o bin/server ./cmd/server && cd ../..; \
	done
	@echo "🔨 Compilando gateway..."
	@cd gateway && go build -o bin/server ./cmd/server && cd ..

# Roda migrations (requer migrate CLI instalado)
migrate:
	@echo "📦 Rodando migrations..."
	@for service in auth-service catalog-service order-service payment-service; do \
		echo "  → $$service"; \
		migrate -path services/$$service/migrations -database "$(DATABASE_URL)" up; \
	done

