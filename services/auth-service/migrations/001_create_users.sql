-- ─────────────────────────────────────────────────────────────
-- Migration: 001_create_users.sql
-- Tabela de usuários do schema auth
-- ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS auth.users (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'customer',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Garante unicidade e buscas em tempo recorde por email
CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_users_email ON auth.users (LOWER(email));

-- Índice para consultas filtrando por permissão/role
CREATE INDEX IF NOT EXISTS idx_auth_users_role ON auth.users (role);

