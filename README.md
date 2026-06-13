# Gaver

> Motor de execução e gerenciador de módulos para projetos de qualquer tipo — do serviço único à rede de microserviços.

Gaver é um meta-framework CLI que unifica scaffolding, execução e orquestração de projetos em uma única interface. Ele não gera código e para: ele também é responsável por rodar, inicializar, compilar e parar os projetos que gera.

Tudo é configurado por um único manifesto: `gaver.json`.

---

## O que o Gaver faz

| Capacidade | O que significa |
|---|---|
| **Scaffolding** | Gera estruturas de projeto a partir de repositórios externos de templates |
| **Execução** | Roda qualquer projeto via `gaver run`, lendo comandos do `gaver.json` |
| **Orquestração** | Gerencia redes de módulos com `depends_on`, health checks e execução paralela |
| **Módulos externos** | Cada módulo é um repositório Git independente com seu próprio `gaver.json` |
| **Installs reproduzíveis** | `gaver.lock` fixa commits exatos de cada módulo instalado |
| **Comunicação entre módulos** | Módulos exportam endpoints e canais; dependentes os recebem como variáveis de ambiente |

---

## Instalação

```sh
go install github.com/Dalistor/gaver/cmd/gaver@latest
```

---

## Início rápido

```sh
# 1. Registre um repositório de templates
gaver repo add oficial https://github.com/Dalistor/gaver-templates

# 2. Crie um projeto
gaver new --type api --name meu-servico

# 3. Entre no projeto e instale dependências
cd meu-servico
gaver init

# 4. Rode
gaver run
```

---

## gaver.json

O manifesto de cada projeto ou módulo:

```json
{
  "name": "meu-projeto",
  "version": "1.0.0",
  "type": "api",
  "platform": "linux",
  "exports": {
    "api": { "protocol": "grpc", "address": "localhost:50051", "schema": "proto/api.proto" }
  },
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
      "name": "payments",
      "source": "https://github.com/minha-org/gaver-payments",
      "depends_on": ["database"],
      "env_from": ["database"],
      "env": { "PAYMENTS_PORT": "3001" },
      "health": {
        "url": "http://localhost:3001/health",
        "timeout": "60s",
        "interval": "3s"
      }
    }
  ],
  "commands": {
    "init": "npm install",
    "run": "npm start",
    "build": "npm run build",
    "migrate": "npm run db:migrate"
  }
}
```

### Campos do manifesto

| Campo | Tipo | Descrição |
|---|---|---|
| `name` | string | Nome do projeto ou módulo (obrigatório) |
| `version` | string | Versão semântica |
| `type` | string | Tipo do projeto (livre: `api`, `worker`, `webapp`, etc.) |
| `platform` | string | Plataforma alvo: `linux`, `darwin`, `windows`, `any` |
| `parent` | string | Nome do módulo pai (usado em sub-módulos) |
| `exports` | object | Canais e endpoints que este módulo expõe aos dependentes |
| `modules` | array | Sub-módulos referenciados |
| `commands` | object | Mapa de comandos: `init`, `run`, `build`, ou qualquer chave customizada |

### Campos de cada módulo em `modules[]`

| Campo | Tipo | Descrição |
|---|---|---|
| `name` | string | Identificador local do módulo (obrigatório) |
| `source` | string | URL do repositório Git (`https://`, `git@`, `ssh://`) |
| `depends_on` | array | Módulos que devem estar prontos antes deste |
| `health` | object | Configuração de health check HTTP |
| `env` | object | Variáveis de ambiente estáticas injetadas neste módulo |
| `env_from` | array | Módulos cujos `exports` serão injetados como variáveis de ambiente |

---

## Comandos

### `gaver repo`

Gerencia repositórios de templates registrados em `~/.gaver/config.json`.

```sh
gaver repo add <name> <url>     # Registra um repositório
gaver repo list                 # Lista repositórios registrados
gaver repo remove <name>        # Remove um repositório
```

**Exemplos:**

```sh
gaver repo add oficial https://github.com/Dalistor/gaver-templates
gaver repo add minha-org git@github.com:minha-org/templates.git
```

---

### `gaver new`

Cria um novo projeto a partir de um repositório de templates.

```sh
gaver new --type <tipo> --name <nome> [--from <repo>]
```

| Flag | Obrigatório | Descrição |
|---|---|---|
| `--type`, `-t` | Sim | Tipo do projeto (`api`, `webapp`, `agent`, etc.) |
| `--name`, `-n` | Sim | Nome do projeto |
| `--from`, `-f` | Não | Repositório de templates a usar (obrigatório se houver mais de um registrado) |

**Exemplos:**

```sh
gaver new --type api --name meu-servico
gaver new --type webapp --name meu-site --from minha-org
```

---

### `gaver install`

Baixa e instala todos os módulos declarados em `gaver.json`, recursivamente. Usa `gaver.lock` para garantir instalações reproduzíveis: se o lock já existir, instala exatamente o commit fixado.

```sh
gaver install
```

Após instalar, um `gaver.lock` é criado ou atualizado com o SHA do commit de cada módulo.

---

### `gaver init`

Executa o comando `commands.init` em cascata — no projeto raiz e em todos os módulos instalados, respeitando a ordem de `depends_on`.

```sh
gaver init [--parallel]
```

| Flag | Descrição |
|---|---|
| `--parallel`, `-p` | Executa módulos independentes em paralelo |

---

### `gaver run`

Executa o projeto. Se houver módulos declarados, inicia a rede inteira em background:
- Respeita `depends_on` para ordem de inicialização
- Aguarda health checks antes de subir o próximo nível
- Bloqueia até Ctrl+C, depois encerra tudo com graceful shutdown

```sh
gaver run [--parallel]
```

| Flag | Descrição |
|---|---|
| `--parallel`, `-p` | Inicia módulos independentes em paralelo |

**Comportamento:**

- **Sem módulos:** executa `commands.run` em foreground
- **Com módulos:** inicia cada módulo em background com prefixo `[nome]` nos logs, aguarda health checks, e bloqueia

---

### `gaver build`

Executa `commands.build` em cascata em todos os módulos, respeitando `depends_on`.

```sh
gaver build [--parallel]
```

---

### `gaver exec`

Executa qualquer comando customizado declarado no `gaver.json` em cascata por toda a hierarquia de módulos.

```sh
gaver exec <command> [--parallel]
```

**Exemplos:**

```sh
gaver exec migrate
gaver exec seed --parallel
gaver exec test
```

Todos os módulos instalados que tiverem `"migrate"` (ou qualquer chave) em seus `commands` serão executados na ordem correta.

---

### `gaver status`

Mostra o status dos módulos atualmente gerenciados pelo Gaver.

```sh
gaver status
```

```
MÓDULO                   PID        STATUS
─────────────────────────────────────────────
api                      12345      rodando
payments                 12346      rodando
database                 12340      rodando
```

---

### `gaver stop`

Para todos os módulos em execução com encerramento gracioso (SIGTERM → aguarda → SIGKILL).

```sh
gaver stop [--timeout <duration>]
```

| Flag | Padrão | Descrição |
|---|---|---|
| `--timeout` | `30s` | Tempo máximo para encerramento antes de SIGKILL |

---

### `gaver gen module`

Gera a estrutura de um novo módulo de domínio dentro do projeto atual.

```sh
gaver gen module --name <nome> [--from <repo>]
```

| Flag | Obrigatório | Descrição |
|---|---|---|
| `--name`, `-n` | Sim | Nome do módulo |
| `--from`, `-f` | Não | Repositório de templates (obrigatório se houver mais de um registrado) |

**Exemplo:**

```sh
gaver gen module --name orders
gaver gen module --name payments --from minha-org
```

---

## Plataformas suportadas

| Plataforma | Valor em `platform` |
|---|---|
| Linux | `linux` |
| macOS | `darwin` ou `macos` |
| Windows | `windows` |
| Qualquer | `any` (padrão se omitido) |

O Gaver valida a plataforma no manifesto antes de executar. Módulos incompatíveis com o SO atual são rejeitados na leitura do `gaver.json`.

---

## Comunicação entre módulos

Módulos em linguagens diferentes se comunicam através de variáveis de ambiente injetadas pelo Gaver. O módulo produtor declara o que expõe; o módulo consumidor declara de quem quer receber. O Gaver faz a fiação automaticamente no momento certo — após o health check do produtor e antes de iniciar o consumidor.

```
database (Go)            api (Python)
exports:                 env_from: ["database"]
  functions: grpc://...  ↓ Gaver injeta:
  stream: unix://...     DATABASE_FUNCTIONS=grpc://localhost:50051
                         DATABASE_STREAM=unix://.gaver/sockets/db.sock
```

Veja o guia completo em [docs/modules.md — Comunicação entre módulos](docs/modules.md#comunicação-entre-módulos).

---

## Skill para AI

Importe o arquivo [`prompts/gaver-module.md`](prompts/gaver-module.md) em qualquer AI para obter um assistente especializado em criar módulos Gaver compatíveis. Funciona com Claude, ChatGPT, Cursor, GitHub Copilot e qualquer ferramenta que aceite instruções de sistema.

Cole o conteúdo do arquivo como:
- **Claude / ChatGPT** — "Custom instructions" ou início do chat
- **Cursor** — `.cursor/rules/gaver.mdc` ou "AI Rules" nas configurações
- **GitHub Copilot** — `.github/copilot-instructions.md`
- **Qualquer outro AI** — início da conversa como system prompt

---

## Como criar um repositório de templates compatível

Um repositório Gaver-compatível deve seguir esta estrutura:

```
projects/
  api/              ← template de projeto tipo "api"
    gaver.json
    ...
  webapp/
    ...
modules/
  crud/             ← template de módulo tipo "crud"
    gaver.json
    ...
```

Veja [docs/modules.md](docs/modules.md) para o guia completo de criação de módulos e sub-módulos.

---

## Autor

Feito por [Dalistor](https://github.com/Dalistor).

## Licença

Distribuído sob a [Licença MIT](LICENSE). Livre para usar, modificar e distribuir — inclusive em projetos comerciais.
