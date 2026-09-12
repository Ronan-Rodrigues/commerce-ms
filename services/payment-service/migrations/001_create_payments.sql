-- ─────────────────────────────────────────────────────────────
-- Migration: 001_create_payments.sql
-- Tabela de transações do schema payments
-- ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS payments.transactions (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status VARCHAR(50) NOT NULL DEFAULT 'succeeded',
    payment_method VARCHAR(50) NOT NULL DEFAULT 'pix',
    gateway_transaction_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments.transactions (order_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments.transactions (status);
CREATE INDEX IF NOT EXISTS idx_payments_gateway_tx ON payments.transactions (gateway_transaction_id);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments.transactions (created_at);

