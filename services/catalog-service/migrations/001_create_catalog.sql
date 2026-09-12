-- ─────────────────────────────────────────────────────────────
-- Migration: 001_create_catalog.sql
-- Tabelas de categorias e produtos do schema catalog
-- ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS catalog.categories (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_catalog_categories_slug ON catalog.categories (slug);

CREATE TABLE IF NOT EXISTS catalog.products (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    price BIGINT NOT NULL CHECK (price > 0), -- Preço em centavos
    stock INT NOT NULL CHECK (stock >= 0) DEFAULT 0,
    sku VARCHAR(100) NOT NULL UNIQUE,
    category_id VARCHAR(36) REFERENCES catalog.categories(id) ON DELETE SET NULL,
    image_url TEXT,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_catalog_products_slug ON catalog.products (slug);
CREATE INDEX IF NOT EXISTS idx_catalog_products_sku ON catalog.products (sku);
CREATE INDEX IF NOT EXISTS idx_catalog_products_category ON catalog.products (category_id);
CREATE INDEX IF NOT EXISTS idx_catalog_products_price ON catalog.products (price);
CREATE INDEX IF NOT EXISTS idx_catalog_products_active ON catalog.products (active);

