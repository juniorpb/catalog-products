-- 1. Criar a nova tabela de Vendedores (Seller)
CREATE TABLE IF NOT EXISTS Seller (
  id TEXT PRIMARY KEY,
  name TEXT UNIQUE NOT NULL
);