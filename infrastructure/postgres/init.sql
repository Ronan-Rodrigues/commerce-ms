-- ─────────────────────────────────────────────────────────────
-- Inicialização do PostgreSQL
-- Cria um schema por serviço — isolamento de domínio
-- Executado automaticamente na primeira vez que o container sobe
-- ─────────────────────────────────────────────────────────────

-- Schema do auth-service
CREATE SCHEMA IF NOT EXISTS auth;

-- Schema do catalog-service
CREATE SCHEMA IF NOT EXISTS catalog;

-- Schema do order-service
CREATE SCHEMA IF NOT EXISTS orders;

-- Schema do payment-service
CREATE SCHEMA IF NOT EXISTS payments;

