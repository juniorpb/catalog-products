# Débitos Técnicos

Mapeamento das falhas técnicas e de negócio identificadas no projeto. Organizado por categoria e criticidade.

---

## Falhas Técnicas

### 1. Carga total em memória — escalabilidade comprometida com alto volume

**Criticidade:** Alta

As três funções de carregamento buscam **todas as linhas** das tabelas sem qualquer filtro ou paginação:

```go
// product.go
SELECT Id, COALESCE(Name, ''), COALESCE(Brand, ''), COALESCE(Category, '') FROM Product

// seller.go
SELECT id, name FROM Seller

// sellerproduct.go
SELECT seller_id, product_id FROM SellerProduct
```

Com 1 milhão de produtos, o processo carrega ~200 MB apenas para o mapa de produtos antes de processar a primeira entrada do JSON. A tabela `SellerProduct` pode crescer N × M (vendedores × produtos), tornando o carregamento inviável em produção.

**Arquivos afetados:**
- `internal/database/product.go` — `LoadAllProducts`
- `internal/database/seller.go` — `LoadAllSellers`
- `internal/database/sellerproduct.go` — `LoadAllSellerProducts`

**Solução sugerida:** Processar o arquivo de entrada em lotes (_batches_) e fazer lookups pontuais no banco por nome normalizado, ou usar uma tabela de staging temporária com JOIN para identificar novidades.

---

### 2. Sem constraint `UNIQUE` em `SellerProduct(external_id)`

**Criticidade:** Média

O `external_id` é o identificador do produto no sistema do vendedor. Não há garantia de unicidade no banco. Dois sellers distintos podem ter coincidentemente o mesmo `external_id`, o que é semanticamente inválido e aceito sem erro.

**Solução sugerida:** Adicionar constraint `UNIQUE (seller_id, external_id)` para garantir que o mesmo identificador externo não seja cadastrado duas vezes para o mesmo vendedor.

---

### 3. Leitura do JSON inteiro em memória

**Criticidade:** Média

`os.ReadFile` carrega o arquivo completo em um `[]byte` antes do parse. Um JSON de 10 GB esgota a memória antes mesmo de iniciar o processamento.

```go
// entry.go
data, err := os.ReadFile(path)
json.Unmarshal(data, &entries)
```

**Solução sugerida:** Usar `json.NewDecoder` com streaming sobre `*os.File`, processando um objeto por vez.

---

### 4. Variável global `database.DB`

**Criticidade:** Média

O estado da conexão é exposto como variável global, dificultando testes, impedindo múltiplas instâncias e tornando a injeção de dependência implícita. O `main.go` acessa `database.DB` diretamente, acoplando o ponto de entrada à implementação interna.

**Solução sugerida:** `ConnectDB` retornar `*sql.DB` em vez de armazenar em variável global. O `main.go` passa a instância como parâmetro para todas as funções que precisam do banco.

---

### 5. `processEntries` não é uma função pura apesar do comentário sugerir

**Criticidade:** Baixa

A função chama `InsertSeller`, `InsertProduct` e `InsertSellerProduct` — todas com efeitos colaterais de banco. Funções com I/O não são puras. Isso tem impacto direto nos testes: qualquer teste de `processEntries` sempre depende de um banco real ou mock.

---

### 6. Migrations executadas sem controle de versão aplicado

**Criticidade:** Baixa agora, Alta no futuro

Todas as migrations rodam a cada inicialização. Funciona agora porque são `IF NOT EXISTS`, mas impede a adição de qualquer migration com operação destrutiva (`ALTER TABLE`, `UPDATE`, `DELETE`). O sistema perdeu a capacidade de evoluir o schema de forma controlada.

**Solução sugerida:** Reintroduzir a tabela `schema_migrations` para rastrear quais migrations já foram aplicadas, permitindo operações destrutivas com segurança.

---

### 7. Deduplicação por `external_id` quebrada após substituição de UUID inválido

**Criticidade:** Média

UUIDs inválidos recebem novos IDs **aleatórios** a cada execução. Em seguida, `deduplicateByExternalID` não consegue detectar que duas entradas com UUIDs originalmente inválidos e iguais são a mesma — porque agora têm IDs distintos gerados aleatoriamente.

```go
// sanitize → gera UUID novo aleatório
e.Id = uuid.New().String()

// deduplicate → nunca vai encontrar duplicata para esses casos
if seen[e.Id] { continue }
```

---

## Falhas de Negócio

### 8. Dados de produtos existentes nunca são atualizados

**Criticidade:** Média

Se um produto já existe no banco com `Brand = NULL` e o JSON traz o mesmo produto com `Brand = "Samsung"`, o sistema detecta que ele existe e retorna o registro antigo **sem enriquecer os dados**.

```go
// ensureProduct — retorna imediatamente sem verificar se há dados novos
if p, ok := products[key]; ok {
    return p, nil
}
```

**Solução sugerida:** Implementar um `UPDATE` condicional quando o registro existente tiver campos nulos que o novo dado pode preencher.

---

### 9. Caminho do arquivo JSON hardcoded

**Criticidade:** Baixa

O sistema só processa um arquivo em um caminho fixo relativo ao diretório de trabalho. Não é possível processar um arquivo diferente sem alterar o código.

```go
jsonPath := filepath.Join(projectRoot(), "data", "ProductEntry.json")
```

**Solução sugerida:** Receber o caminho via argumento de linha de comando (`os.Args`) ou variável de ambiente.

---

### 10. Sem relatório de entradas rejeitadas ou problemáticas

**Criticidade:** Média

Entradas descartadas por `Name` vazio, `SellerName` vazio ou UUID inválido não geram nenhum log específico. Em produção, o operador não sabe quais registros do vendedor foram problemáticos nem em qual linha do arquivo estavam.

**Solução sugerida:** Retornar junto com `ConsolidationStats` uma lista de `RejectedEntry{Line int, Reason string, Original ProductEntry}` e logar ou persistir esse relatório ao final da execução.

---

## Resumo por Criticidade

| # | Problema | Categoria | Criticidade |
|---|---|---|---|
| 1 | Carga total em memória | Técnica | Alta |
| 2 | Sem UNIQUE em SellerProduct(external_id) | Técnica | Média |
| 3 | JSON carregado inteiro em memória | Técnica | Média |
| 4 | Variável global database.DB | Técnica | Média |
| 7 | Dedup quebrado após substituição de UUID inválido | Técnica | Média |
| 8 | Dados existentes nunca são atualizados | Negócio | Média |
| 10 | Sem relatório de entradas rejeitadas | Negócio | Média |
| 5 | processEntries não é pura apesar do comentário | Técnica | Baixa |
| 6 | Migrations sem controle de versão | Técnica | Baixa agora / Alta no futuro |
| 9 | Caminho do JSON hardcoded | Negócio | Baixa |
