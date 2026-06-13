# Criando módulos e sub-módulos no Gaver

O Gaver trata cada módulo como um projeto autônomo: ele tem seu próprio `gaver.json`, seus próprios comandos e pode declarar os sub-módulos dos quais depende. Um projeto pode ser uma aplicação inteira; uma rede de módulos pode ser uma infraestrutura completa.

---

## O que é um módulo

No Gaver, um módulo é qualquer coisa que tenha um `gaver.json` e saiba como executar a si mesmo. Pode ser:

- Um microserviço (API REST, worker, gRPC)
- Um banco de dados com migrações
- Uma fila de mensagens
- Um frontend
- Um job batch
- Qualquer combinação dos acima

A hierarquia não tem limite fixo de profundidade: um projeto instala módulos, cada módulo pode instalar sub-módulos, e assim por diante até 20 níveis.

---

## Estrutura de um módulo

Todo módulo é um repositório Git com um `gaver.json` na raiz:

```
meu-modulo/
├── gaver.json          ← manifesto do módulo
├── modules/            ← sub-módulos instalados (criado pelo gaver install)
│   └── database/
│       └── gaver.json
├── src/
│   └── ...
└── gaver.lock          ← commits fixados dos sub-módulos
```

### gaver.json mínimo

```json
{
  "name": "payments",
  "version": "1.0.0",
  "type": "service",
  "commands": {
    "init": "npm install",
    "run": "node src/index.js",
    "build": "npm run build"
  }
}
```

### gaver.json com sub-módulos

```json
{
  "name": "payments",
  "version": "1.0.0",
  "type": "service",
  "platform": "linux",
  "parent": "meu-projeto",
  "modules": [
    {
      "name": "database",
      "source": "https://github.com/minha-org/gaver-postgres",
      "health": {
        "url": "http://localhost:5432/health",
        "timeout": "60s",
        "interval": "3s"
      }
    },
    {
      "name": "cache",
      "source": "https://github.com/minha-org/gaver-redis",
      "depends_on": ["database"],
      "health": {
        "url": "http://localhost:6379/ping",
        "timeout": "30s",
        "interval": "2s"
      }
    }
  ],
  "commands": {
    "init": "npm install",
    "run": "node src/index.js",
    "build": "npm run build",
    "migrate": "node src/migrate.js"
  }
}
```

---

## Campos do gaver.json

### `modules[]`

Array de referências a sub-módulos. Cada entrada:

| Campo | Tipo | Descrição |
|---|---|---|
| `name` | string | Identificador local do módulo (obrigatório) |
| `source` | string | URL do repositório Git (`https://`, `git@`, `ssh://`) |
| `depends_on` | array | Nomes de módulos que devem estar prontos antes deste |
| `health` | object | Configuração de health check HTTP |
| `env` | object | Variáveis de ambiente estáticas injetadas neste módulo |
| `env_from` | array | Módulos cujos `exports` serão injetados como variáveis de ambiente |

### `health`

Antes de subir o próximo nível de módulos, o Gaver aguarda o health check do módulo atual:

| Campo | Tipo | Padrão | Descrição |
|---|---|---|---|
| `url` | string | — | Endpoint HTTP que retorna 2xx quando pronto |
| `timeout` | string | `30s` | Tempo máximo total de espera |
| `interval` | string | `2s` | Intervalo entre tentativas |

```json
"health": {
  "url": "http://localhost:8080/health",
  "timeout": "60s",
  "interval": "3s"
}
```

### `exports`

Canais e endpoints que este módulo expõe. Cada chave é o nome do export:

| Campo | Tipo | Descrição |
|---|---|---|
| `protocol` | string | Protocolo do canal: `grpc`, `http`, `unix`, `amqp`, `tcp`, etc. |
| `address` | string | Endereço de conexão: `host:porta` ou caminho de socket |
| `schema` | string | Opcional — caminho para o contrato da interface (`.proto`, `openapi.yaml`, etc.) |

```json
"exports": {
  "functions": {
    "protocol": "grpc",
    "address": "localhost:50051",
    "schema": "proto/payments.proto"
  },
  "stream": {
    "protocol": "unix",
    "address": ".gaver/sockets/payments.sock"
  },
  "events": {
    "protocol": "amqp",
    "address": "localhost:5672"
  }
}
```

---

## Criando um módulo do zero

### 1. Crie o repositório Git

```sh
mkdir gaver-payments && cd gaver-payments
git init
```

### 2. Crie o gaver.json

```json
{
  "name": "payments",
  "version": "1.0.0",
  "type": "service",
  "commands": {
    "init": "npm install",
    "run": "node src/index.js",
    "build": "npm run build"
  }
}
```

### 3. Publique

```sh
git add . && git commit -m "feat: módulo payments"
git remote add origin https://github.com/minha-org/gaver-payments
git push -u origin main
```

### 4. Declare no projeto pai

No `gaver.json` do projeto ou módulo pai, adicione a referência:

```json
{
  "name": "meu-projeto",
  "modules": [
    {
      "name": "payments",
      "source": "https://github.com/minha-org/gaver-payments"
    }
  ]
}
```

### 5. Instale

```sh
gaver install
```

O Gaver clona o repositório em `modules/payments/` e registra o commit no `gaver.lock`.

---

## Usando `depends_on`

`depends_on` define a ordem de inicialização. O Gaver calcula os níveis automaticamente (topological sort) e executa cada nível antes de avançar.

```json
"modules": [
  { "name": "database", "source": "..." },
  { "name": "cache",    "source": "...", "depends_on": ["database"] },
  { "name": "api",      "source": "...", "depends_on": ["database", "cache"] },
  { "name": "worker",   "source": "...", "depends_on": ["database", "cache"] }
]
```

Ordem de execução resultante:

```
Nível 0: database
Nível 1: cache          ← aguarda database ficar pronto (health check)
Nível 2: api, worker    ← executados em paralelo (ambos dependem só do nível 1)
```

Ciclos em `depends_on` são detectados na inicialização e retornam erro com os módulos envolvidos.

---

## Gerando um módulo com template

Se o repositório de templates tiver um diretório `modules/`, você pode gerar a estrutura inicial de um módulo com:

```sh
gaver gen module --name orders
gaver gen module --name users --from minha-org
```

O módulo é gerado em `src/modules/<nome>/`. A estrutura gerada depende do repositório de templates configurado.

---

## Hierarquia de projetos

O Gaver não impõe limite sobre o que um módulo pode conter. Uma hierarquia típica de rede completa:

```
meu-sistema/                    ← projeto raiz
├── gaver.json                  ← declara: api, workers, infra
├── gaver.lock
└── modules/
    ├── api/                    ← módulo "api"
    │   ├── gaver.json          ← declara: database, cache
    │   └── modules/
    │       ├── database/
    │       └── cache/
    ├── workers/                ← módulo "workers"
    │   ├── gaver.json          ← declara: queue
    │   └── modules/
    │       └── queue/
    └── infra/                  ← módulo "infra"
        └── gaver.json
```

Ao rodar `gaver run` na raiz, o Gaver:
1. Lê o `gaver.json` raiz
2. Instala e sobe `infra` primeiro (sem dependências)
3. Aguarda health checks de `infra`
4. Sobe `api` e `workers` em paralelo (se `--parallel`)
5. Dentro de cada um, repete o processo com seus próprios sub-módulos

---

## Comandos customizados

Além de `init`, `run` e `build`, qualquer chave em `commands` pode ser executada com `gaver exec`:

```json
"commands": {
  "migrate": "node scripts/migrate.js",
  "seed":    "node scripts/seed.js",
  "test":    "npm test"
}
```

```sh
gaver exec migrate           # executa em todos os módulos, em ordem
gaver exec seed --parallel   # executa nos módulos independentes em paralelo
```

O Gaver propaga o comando por toda a hierarquia, executando apenas nos módulos que tiverem aquela chave declarada.

---

## Comunicação entre módulos

O Gaver fornece uma camada de fiação automática entre módulos: o módulo produtor declara o que expõe (`exports`) e o módulo consumidor declara de onde quer receber (`env_from`). O Gaver resolve tudo em tempo de boot — não há proxy, não há sidecar, não há overhead em runtime.

### Como funciona o fluxo completo

```
gaver run
  │
  ├─ Nível 0: sobe "database"
  │     ├─ inicia processo com env injetado pelo pai
  │     ├─ aguarda health check: GET http://localhost:5432/health → 200 OK
  │     └─ salva exports em .gaver/exports/database.json
  │
  ├─ Nível 1: sobe "cache" e "api" (independentes entre si)
  │     ├─ lê .gaver/exports/database.json
  │     ├─ converte exports → variáveis de ambiente
  │     ├─ inicia processos com essas vars + env estático do ModuleRef
  │     ├─ aguarda health checks
  │     └─ salva exports de "cache" e "api"
  │
  └─ Rede pronta. Ctrl+C → graceful shutdown.
```

As variáveis de ambiente ficam disponíveis para o processo do módulo consumidor da mesma forma que qualquer env var do sistema operacional — sem configuração adicional no código do módulo.

---

### Exportando um canal

No `gaver.json` do módulo produtor, declare o campo `exports`:

```json
{
  "name": "database",
  "version": "1.0.0",
  "exports": {
    "functions": {
      "protocol": "grpc",
      "address": "localhost:50051",
      "schema": "proto/database.proto"
    },
    "stream": {
      "protocol": "unix",
      "address": ".gaver/sockets/database.sock"
    }
  },
  "commands": {
    "run": "./database-server"
  }
}
```

Um módulo pode exportar quantos canais quiser com nomes arbitrários. O nome do export (`"functions"`, `"stream"`) vira parte do nome da variável de ambiente.

---

### Consumindo exports de outro módulo

No `gaver.json` do projeto pai, declare `env_from` na referência ao módulo consumidor:

```json
{
  "name": "meu-sistema",
  "modules": [
    {
      "name": "database",
      "source": "https://github.com/minha-org/gaver-database"
    },
    {
      "name": "api",
      "source": "https://github.com/minha-org/gaver-api",
      "depends_on": ["database"],
      "env_from": ["database"]
    }
  ]
}
```

Com isso, quando o módulo `api` iniciar, ele receberá automaticamente:

```
DATABASE_FUNCTIONS=grpc://localhost:50051
DATABASE_FUNCTIONS_SCHEMA=proto/database.proto
DATABASE_STREAM=unix://.gaver/sockets/database.sock
```

---

### Convenção de nomes das variáveis

O Gaver gera os nomes das variáveis a partir do nome do módulo e do nome do export:

```
{NOME_DO_MÓDULO}_{NOME_DO_EXPORT} = {protocol}://{address}
{NOME_DO_MÓDULO}_{NOME_DO_EXPORT}_SCHEMA = {schema}   (se declarado)
```

Hífens e pontos nos nomes são convertidos para `_`. Tudo em maiúsculas.

| Módulo | Export | Variável gerada |
|---|---|---|
| `database` | `functions` | `DATABASE_FUNCTIONS=grpc://localhost:50051` |
| `database` | `stream` | `DATABASE_STREAM=unix://.gaver/sockets/db.sock` |
| `my-cache` | `events` | `MY_CACHE_EVENTS=amqp://localhost:5672` |
| `payments` | `api` | `PAYMENTS_API=http://localhost:3001` |

---

### Injetando variáveis estáticas com `env`

Além de `env_from`, o pai pode injetar variáveis de ambiente fixas diretamente na referência do módulo, sem precisar de exports:

```json
{
  "name": "api",
  "source": "https://...",
  "env": {
    "NODE_ENV": "production",
    "PORT": "8080",
    "LOG_LEVEL": "info"
  }
}
```

`env` e `env_from` podem ser usados juntos — o módulo receberá ambos.

---

### Exemplo completo: rede com três linguagens

Cenário: banco de dados em Go, serviço de pagamentos em Python, API em Node.js.

**`gaver.json` do banco de dados (Go):**

```json
{
  "name": "database",
  "exports": {
    "query": {
      "protocol": "grpc",
      "address": "localhost:50051",
      "schema": "proto/database.proto"
    }
  },
  "health": {
    "url": "http://localhost:50052/health"
  },
  "commands": {
    "run": "./bin/database-server"
  }
}
```

**`gaver.json` do serviço de pagamentos (Python):**

```json
{
  "name": "payments",
  "exports": {
    "api": {
      "protocol": "http",
      "address": "localhost:3001",
      "schema": "openapi/payments.yaml"
    },
    "events": {
      "protocol": "amqp",
      "address": "localhost:5672"
    }
  },
  "health": {
    "url": "http://localhost:3001/health"
  },
  "commands": {
    "init": "pip install -r requirements.txt",
    "run": "python src/main.py"
  }
}
```

O módulo `payments` em Python lê `DATABASE_QUERY` do ambiente para chamar o banco:

```python
import os
import grpc
from proto import database_pb2_grpc

channel = grpc.insecure_channel(
    os.environ["DATABASE_QUERY"].replace("grpc://", "")
)
stub = database_pb2_grpc.DatabaseStub(channel)
```

**`gaver.json` da API (Node.js):**

```json
{
  "name": "api",
  "exports": {
    "gateway": {
      "protocol": "http",
      "address": "localhost:8080"
    }
  },
  "health": {
    "url": "http://localhost:8080/health"
  },
  "commands": {
    "init": "npm install",
    "run": "node src/server.js",
    "build": "npm run build"
  }
}
```

**`gaver.json` do projeto raiz:**

```json
{
  "name": "meu-sistema",
  "version": "1.0.0",
  "modules": [
    {
      "name": "database",
      "source": "https://github.com/minha-org/gaver-database",
      "health": {
        "url": "http://localhost:50052/health",
        "timeout": "60s",
        "interval": "2s"
      }
    },
    {
      "name": "payments",
      "source": "https://github.com/minha-org/gaver-payments",
      "depends_on": ["database"],
      "env_from": ["database"],
      "env": { "PAYMENTS_ENV": "production" },
      "health": {
        "url": "http://localhost:3001/health",
        "timeout": "30s",
        "interval": "3s"
      }
    },
    {
      "name": "api",
      "source": "https://github.com/minha-org/gaver-api",
      "depends_on": ["payments"],
      "env_from": ["database", "payments"],
      "env": { "PORT": "8080" },
      "health": {
        "url": "http://localhost:8080/health",
        "timeout": "30s",
        "interval": "2s"
      }
    }
  ]
}
```

O módulo `api` em Node.js recebe todas as variáveis necessárias sem nenhuma configuração extra:

```js
// process.env disponível automaticamente
const dbChannel = process.env.DATABASE_QUERY   // grpc://localhost:50051
const paymentsUrl = process.env.PAYMENTS_API   // http://localhost:3001
const eventsUrl = process.env.PAYMENTS_EVENTS  // amqp://localhost:5672
```

Ao rodar `gaver run --parallel`, a ordem de inicialização é:

```
1. database   → health check ok → exports salvos
2. payments   → recebe DATABASE_QUERY → health check ok → exports salvos
3. api        → recebe DATABASE_QUERY + PAYMENTS_API + PAYMENTS_EVENTS → pronto
```

---

### Protocolos suportados

O campo `protocol` é livre — o Gaver não interpreta nem valida o protocolo, apenas o usa para montar a URL. Qualquer protocolo que a linguagem do módulo suporte funciona:

| Protocolo | Caso de uso | Exemplo de address |
|---|---|---|
| `grpc` | RPC tipado, funções, streaming bidirecional | `localhost:50051` |
| `http` | REST APIs, webhooks, funções simples | `localhost:8080` |
| `https` | REST APIs com TLS | `localhost:8443` |
| `unix` | Comunicação local de alta performance via socket | `.gaver/sockets/db.sock` |
| `amqp` | Filas de mensagem (RabbitMQ, etc.) | `localhost:5672` |
| `nats` | Pub/sub de baixa latência | `localhost:4222` |
| `tcp` | Protocolo binário customizado | `localhost:9000` |
| `redis` | Cache, pub/sub, streams | `localhost:6379` |

---

### Onde os exports ficam armazenados

O Gaver persiste os exports de cada módulo em `.gaver/exports/<nome>.json` na raiz do projeto. Esses arquivos são criados em runtime após o health check de cada módulo e podem ser inspecionados manualmente:

```
.gaver/
└── exports/
    ├── database.json
    ├── payments.json
    └── api.json
```

Conteúdo de `.gaver/exports/database.json`:

```json
{
  "query": {
    "protocol": "grpc",
    "address": "localhost:50051",
    "schema": "proto/database.proto"
  }
}
```

Esses arquivos **não devem ser commitados** — eles são gerados a cada `gaver run`. Adicione ao `.gitignore`:

```
.gaver/exports/
.gaver/pids/
```

---

## gaver.lock

O `gaver.lock` garante que `gaver install` sempre instale exatamente as mesmas versões:

```json
{
  "locked_at": "2026-06-12T14:30:00Z",
  "modules": {
    "database": {
      "source": "https://github.com/minha-org/gaver-postgres",
      "commit": "a1b2c3d4e5f6..."
    },
    "cache": {
      "source": "https://github.com/minha-org/gaver-redis",
      "commit": "f6e5d4c3b2a1..."
    }
  }
}
```

Commite o `gaver.lock` junto com o `gaver.json` para garantir reprodutibilidade nos outros ambientes.

---

## Skill para AI

Para gerar módulos Gaver com qualquer AI assistente, importe o prompt em [`prompts/gaver-module.md`](../prompts/gaver-module.md). Ele contém o schema completo do `gaver.json`, os padrões de exports/imports, exemplos em múltiplas linguagens e as regras de validação.

Como usar:

| Ferramenta | Como importar |
|---|---|
| **Claude** (claude.ai) | Cole o conteúdo em "Custom instructions" ou no início do chat |
| **Cursor** | Salve como `.cursor/rules/gaver.mdc` no projeto |
| **GitHub Copilot** | Salve como `.github/copilot-instructions.md` |
| **ChatGPT** | Cole em "Customize ChatGPT → Custom instructions" |
| **Qualquer outro** | Cole no início da conversa como system prompt |

---

## Criando um repositório de templates compatível

Para que seus templates sejam usados com `gaver new` e `gaver gen module`, o repositório deve ter esta estrutura:

```
gaver-templates/
├── projects/
│   ├── api/              ← template para --type api
│   │   ├── gaver.json    ← será copiado para o projeto
│   │   └── ...
│   └── webapp/
│       └── ...
└── modules/
    ├── crud/             ← template de módulo "crud"
    │   └── ...
    └── auth/
        └── ...
```

Registre o repositório no Gaver:

```sh
gaver repo add minha-org https://github.com/minha-org/gaver-templates
```

A partir daí, `gaver new --type api --from minha-org` clona o repositório e copia `projects/api/` para o nome do projeto.
