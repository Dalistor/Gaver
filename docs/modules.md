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
