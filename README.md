# Catalog Products

Sistema de consolidação de catálogo para marketplaces. Recebe arquivos de produtos de diferentes vendedores e os integra ao catálogo central, evitando duplicatas e registrando o vínculo entre produto e vendedor.

---

## O Problema

Uma loja que passa a operar como marketplace recebe catálogos de múltiplos vendedores. O mesmo produto pode ser cadastrado por vários vendedores com pequenas variações de texto (espaços extras, capitalização diferente). Duplicar produtos no catálogo é indesejável — o sistema deve reconhecer produtos já existentes e apenas criar o vínculo com o novo vendedor.

---

## Como Executar

**Pré-requisitos:** Go 1.21+ e GCC (necessário para o driver SQLite com CGO).

```bash
make setup   # instala as dependências
make run     # conecta ao banco, roda migrations e consolida o catálogo
make test    # executa todos os testes com output detalhado
```

---

## Estrutura do Projeto

```
cmd/
  main.go                        — ponto de entrada

internal/
  business/
    catalog/
      consolidate.go             — serviço principal de consolidação
      entry.go                   — struct ProductEntry + parsing do JSON
      consolidate_test.go        — testes de processEntries
      sanitize_test.go           — testes de sanitizeEntries
      deduplicate_test.go        — testes de deduplicateByExternalID

  domain/
    model.go                     — structs de domínio: Product, Seller, SellerProduct

  database/
    database.go                  — conexão e execução de migrations
    product.go                   — queries SQL para Product
    seller.go                    — queries SQL para Seller
    sellerproduct.go             — queries SQL para SellerProduct

    migrations/
      v1_create_salles.sql       — cria a tabela Seller
      v2_drop_seller_columns.sql — recria SellerProduct com o novo schema

  foundation/
    files/files.go               — leitura de arquivos .sql de um diretório
    normalize/normalize.go       — normalização de strings e validação de UUID

data/
  catalog.db                     — banco SQLite
  ProductEntry.json              — catálogo de entrada dos vendedores
```

---

## Modelagem de Dados

### Tabelas existentes (banco original)

```sql
Product (
  Id       INTEGER PRIMARY KEY AUTOINCREMENT,
  Name     TEXT NOT NULL,
  Brand    TEXT,
  Category TEXT
)
```

### Tabelas criadas pelas migrations

```sql
-- v1: vendedores
Seller (
  id   TEXT PRIMARY KEY,  -- UUID gerado pela aplicação
  name TEXT UNIQUE NOT NULL
)

-- v2: vínculo produto ↔ vendedor
SellerProduct (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  seller_id   TEXT    NOT NULL,  -- FK → Seller.id
  product_id  INTEGER NOT NULL,  -- FK → Product.Id
  external_id TEXT    NOT NULL   -- UUID original do produto no sistema do vendedor
)
```

### Diagrama de relacionamento

```
Product (1) ←── SellerProduct (N) ───→ (N) Seller
```

Um produto pode ter múltiplos vendedores. Um vendedor pode oferecer múltiplos produtos. `SellerProduct` é a tabela de junção que registra o `external_id` — o identificador que o vendedor usa em seu próprio sistema.

---

## Migrations

As migrations são arquivos `.sql` lidos automaticamente da pasta `internal/database/migrations/`, ordenados alfabeticamente e executados a cada inicialização. Todas são **idempotentes** (`CREATE TABLE IF NOT EXISTS`), portanto seguras para re-execução.

| Arquivo | Descrição |
|---|---|
| `v1_create_salles.sql` | Cria a tabela `Seller` |
| `v2_drop_seller_columns.sql` | Cria a tabela `SellerProduct` com o schema final |

---

## Fluxo de Consolidação

```
1. Ler ProductEntry.json
2. Sanitizar entradas
      └─ trim de espaços em Name, SellerName, Category
      └─ descartar entradas com Name ou SellerName vazio
      └─ substituir UUIDs inválidos por novos UUIDs gerados
3. Deduplicar por external_id (manter primeira ocorrência)
4. Carregar em memória:
      └─ map[normalizedName]Product   (lookup O(1))
      └─ map[normalizedName]Seller    (lookup O(1))
      └─ map[sellerID:productID]bool  (lookup O(1))
5. Para cada entrada:
      a. Verificar se Seller existe → inserir se não existir
      b. Verificar se Product existe (por nome normalizado) → inserir se não existir
      c. Verificar se o vínculo SellerProduct já existe → inserir se não existir
6. Exibir ConsolidationStats ao final
```

A estratégia de carregar os três mapas antes do loop evita N+1 queries ao banco — toda decisão de inserção é tomada com lookups O(1) em memória.

---

## Validações e Proteções

| Cenário | Tratamento |
|---|---|
| UUID inválido no JSON | Substituído por `uuid.New()` — a entrada não é descartada |
| UUID duplicado no mesmo arquivo | Segunda ocorrência ignorada (`deduplicateByExternalID`) |
| Produto com nome similar (`"Widget  Pro"` vs `"Widget Pro"`) | `normalize.String` colapsa espaços e aplica lowercase antes da comparação |
| `Brand` nula no JSON | Mapeado como `*string` e inserido como `NULL` no banco |
| Nome ou SellerName vazio | Entrada descartada em `sanitizeEntries` |
| SQL injection (ex.: `TestBrand'; SELECT 1; --`) | Todas as queries usam `?` com parâmetros — o driver escapa automaticamente |
| Re-execução do programa | Idempotente: migrations usam `IF NOT EXISTS`, vínculos existentes são pulados |

---

## Testes

```bash
make test
```

| Arquivo | Função testada | Casos |
|---|---|---|
| `consolidate_test.go` | `processEntries` | novo produto+seller, link duplicado, produto existente+novo seller, dois registros mesmo seller/produto, erro de DB |
| `sanitize_test.go` | `sanitizeEntries` | UUID válido preservado, UUID inválido substituído, Name vazio descartado, SellerName vazio descartado, trim de espaços, múltiplos casos combinados |
| `deduplicate_test.go` | `deduplicateByExternalID` | sem duplicatas, um duplicado, todos duplicados, lista vazia, preservação da primeira ocorrência |

Os testes de `processEntries` usam um banco SQLite em memória (`:memory:`) para garantir isolamento sem dependência de arquivos externos.

---

## Dependências

| Pacote | Versão | Uso |
|---|---|---|
| `github.com/mattn/go-sqlite3` | v1.14.41 | Driver SQLite (CGO) |
| `github.com/google/uuid` | v1.6.0 | Geração de UUIDs para Sellers e IDs inválidos |

---

## Sugestões de Melhorias Futuras

### Funcionalidades

- **Deduplicação por similaridade semântica** — além de normalizar espaços, aplicar distância de Levenshtein ou embeddings para identificar produtos semanticamente iguais com nomes muito distintos.
- **Suporte a múltiplos arquivos de entrada** — processar um diretório inteiro de JSONs de diferentes vendedores em vez de um único arquivo fixo.
- **API HTTP** — expor endpoints REST para receber catálogos via upload e consultar o estado do catálogo consolidado.
- **Relatório de consolidação** — gerar um arquivo de saída com os produtos inseridos, duplicatas encontradas e vínculos criados por execução.

### Qualidade e Operação

- **Rastreamento de migrations com `schema_migrations`** — adicionar controle de quais migrations já foram aplicadas para suportar migrations destrutivas com segurança em produção.
- **Transação por lote** — envolver `processEntries` em uma única transação SQL para garantir atomicidade: ou tudo é persistido ou nada é.
- **Logging estruturado** — substituir `log.Printf` por `slog` (Go 1.21+) com campos estruturados (`zerolog` ou `zap` como alternativas), facilitando integração com sistemas de observabilidade.
- **Configuração por variáveis de ambiente** — externalizar caminhos do banco e do arquivo JSON via `os.Getenv` ou biblioteca como `github.com/caarlos0/env`.
- **Testes de integração** — adicionar testes end-to-end que executam `Consolidate` completo contra um banco `:memory:` com dados pré-populados e verificam o estado final.
- **Lint e CI** — configurar `golangci-lint` e um pipeline de CI (GitHub Actions) com build, testes e lint automatizados.
