CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pgcrypto;


CREATE TABLE IF NOT EXISTS clients (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name        TEXT NOT NULL,
  email       TEXT,
  phone       TEXT,
  meta        JSONB NOT NULL DEFAULT '{}'::jsonb,  
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (name, email)
);

CREATE INDEX IF NOT EXISTS idx_clients_meta_gin ON clients USING GIN (meta);


CREATE TABLE IF NOT EXISTS quotes (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  client_id    UUID REFERENCES clients(id) ON DELETE SET NULL,
  items        JSONB NOT NULL DEFAULT '[]',         
  labor_hours  NUMERIC(10,2) NOT NULL DEFAULT 0,
  labor_rate   NUMERIC(10,2) NOT NULL DEFAULT 0,
  margin_pct   NUMERIC(5,2)  NOT NULL DEFAULT 0,    
  tax_pct      NUMERIC(5,2)  NOT NULL DEFAULT 0,
  subtotal     NUMERIC(12,2) NOT NULL DEFAULT 0,
  total        NUMERIC(12,2) NOT NULL DEFAULT 0,
  currency     TEXT NOT NULL DEFAULT 'USD',
  notes        TEXT,
  public_id    TEXT UNIQUE,                         
  status       TEXT NOT NULL DEFAULT 'draft',       
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE quotes
  ADD CONSTRAINT chk_margin_pct_range CHECK (margin_pct BETWEEN 0 AND 100),
  ADD CONSTRAINT chk_tax_pct_range    CHECK (tax_pct BETWEEN 0 AND 100),
  ADD CONSTRAINT chk_totals_nonneg    CHECK (subtotal >= 0 AND total >= 0);

CREATE INDEX IF NOT EXISTS idx_quotes_client  ON quotes(client_id);
CREATE INDEX IF NOT EXISTS idx_quotes_status  ON quotes(status);
CREATE INDEX IF NOT EXISTS idx_quotes_created ON quotes(created_at DESC);
