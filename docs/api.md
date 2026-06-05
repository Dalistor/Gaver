# Tipo: API

Documentação do tipo `api` gerado pelo Gaver. Este tipo cria uma API HTTP em Go com engine próprio, módulos de domínio e infraestrutura para agentes IA.

---

## Criando um projeto

```sh
gaver new --type api --name meu-servico
gaver new --type api --name meu-servico --database postgres
```

Após criar, acesse o diretório e inicialize:

```sh
cd meu-servico
gaver init   # instala dependências
gaver run    # inicia o servidor
```

---

## Estrutura gerada

```
meu-servico/
├── gaver.json                  ← manifesto do projeto
├── go.mod
├── .env                        ← variáveis de ambiente (não versionado)
├── .gitignore
├── main.go                     ← entrada: registra módulos no engine
├── README.md
├── src/
│   ├── engine/
│   │   └── engine.go           ← motor HTTP, conexão com banco e registro de módulos
│   └── modules/
│       └── example/
│           └── module.go       ← módulo de exemplo
└── ai/
    ├── .pipeline/              ← gerado pelo Gaver, não editar arquivos _
    │   ├── _context.md
    │   ├── _instructions.md
    │   ├── _structure.md
    │   └── custom/             ← regras customizadas para agentes
    └── .memory/
        ├── _status.md          ← resumo do estado atual do projeto
        └── implements/
            └── 0001.md         ← histórico sequencial de implementações
```

---

## `gaver.json`

Manifesto central do projeto. O Gaver lê este arquivo para saber como inicializar, executar e compilar o projeto.

```json
{
  "name": "meu-servico",
  "version": "1.0.0",
  "type": "api",
  "database": {
    "type": "postgres"
  },
  "commands": {
    "init": "go mod tidy",
    "run": "go run ./main.go",
    "build": "go build -o bin/app ."
  }
}
```

O campo `commands` é executado diretamente pelo Gaver — adapte conforme necessário para o seu ambiente.

---

## Engine

O engine é o motor central da API. Ele gerencia:

- Registro de módulos
- Roteamento HTTP
- Conexão com banco de dados (quando configurado)

### Interface de módulo

Todo módulo deve implementar a interface definida em `src/engine/engine.go`:

```go
type Module interface {
    Name() string
    Setup(e *Engine) error
}
```

### Recursos expostos pelo engine

```go
e.Handle(pattern, handler)  // registra uma rota HTTP
e.DB()                      // retorna a conexão com o banco (nil se não configurado)
```

---

## Módulos

Cada módulo encapsula uma seção da regra de negócio. Um módulo pode interagir com múltiplas tabelas — o critério de separação é o domínio, não a tabela.

### Criando um módulo

1. Crie a pasta `src/modules/<nome>/`
2. Crie o arquivo `module.go` implementando a interface `engine.Module`
3. Registre no `main.go`

**Exemplo — `src/modules/orders/module.go`:**

```go
package orders

import (
    "net/http"
    "meu-servico/src/engine"
)

type Module struct{}

func New() *Module {
    return &Module{}
}

func (m *Module) Name() string {
    return "orders"
}

func (m *Module) Setup(e *engine.Engine) error {
    e.Handle("/orders", http.HandlerFunc(m.list))
    e.Handle("/orders/create", http.HandlerFunc(m.create))
    return nil
}

func (m *Module) list(w http.ResponseWriter, r *http.Request) {
    // lógica de listagem
}

func (m *Module) create(w http.ResponseWriter, r *http.Request) {
    // lógica de criação
}
```

### Registrando no `main.go`

```go
package main

import (
    "meu-servico/src/engine"
    "meu-servico/src/modules/orders"
    "meu-servico/src/modules/users"
)

func main() {
    e := engine.New()
    e.Register(orders.New())
    e.Register(users.New())
    e.Start()
}
```

O registro é sempre explícito — o engine não descobre módulos automaticamente.

---

## Banco de dados

O banco é configurado via flag `--database` no `gaver new` e fica registrado no `gaver.json`. A conexão é inicializada automaticamente pelo engine na chamada de `e.Start()`.

| Banco | Driver | Variável de ambiente |
|---|---|---|
| `postgres` | `github.com/lib/pq` | `DB_DSN` |
| `mysql` | `github.com/go-sql-driver/mysql` | `DB_DSN` |
| `sqlite` | `github.com/mattn/go-sqlite3` | `DB_PATH` |

Configure no `.env`:

```sh
# postgres
DB_DSN=postgres://user:password@localhost:5432/meu-servico?sslmode=disable

# mysql
DB_DSN=user:password@tcp(localhost:3306)/meu-servico

# sqlite
DB_PATH=./meu-servico.db
```

Dentro de um módulo, acesse a conexão via `e.DB()`:

```go
func (m *Module) Setup(e *engine.Engine) error {
    db := e.DB()
    // use db para queries
    return nil
}
```

---

## Infraestrutura IA

### Pipeline

Os arquivos em `ai/.pipeline/` são gerados pelo Gaver e fornecem contexto para agentes IA operarem no projeto:

| Arquivo | Conteúdo | Editável |
|---|---|---|
| `_context.md` | Escopo e limites de atuação do agente | Não |
| `_instructions.md` | Fluxo de trabalho e regras de contexto | Não |
| `_structure.md` | Mapa completo dos arquivos do projeto | Não |
| `custom/*.md` | Regras customizadas do projeto | Sim |

Para adicionar regras específicas do seu projeto (convenções, padrões de código, restrições), crie arquivos `.md` em `ai/.pipeline/custom/`.

### Memória

O histórico de implementações fica em `ai/.memory/implements/`. Cada entrada é um arquivo numerado sequencialmente:

```
ai/.memory/
├── _status.md          ← resumo do estado atual (lido primeiro por agentes)
└── implements/
    ├── 0001.md         ← estrutura inicial gerada pelo Gaver
    ├── 0002.md         ← primeira implementação
    └── 0003.md         ← segunda implementação
```

Ao final de cada implementação, documente em `implements/XXXX.md` e atualize o `_status.md` com o estado atual do projeto. Isso garante que qualquer agente possa retomar o trabalho sem precisar ler todo o histórico.
