# Gaver

> Um meta-framework versátil para geração e execução de estruturas de trabalho modulares em múltiplas plataformas — com infraestrutura nativa para agentes IA.

Gaver unifica scaffolding de projetos, runtime de execução e integração com agentes IA em uma única interface consistente. Ele não apenas gera a estrutura do projeto — ele também é o responsável por rodá-lo. Tudo configurado por um único manifesto: `gaver.json`.

---

## O que o Gaver faz

A maioria das ferramentas de scaffolding gera uma estrutura de pastas e para. O Gaver vai além: ele gera, executa e mantém o projeto, além de fornecer infraestrutura pronta para agentes IA trabalharem dentro dele com contexto preciso e memória persistente.

| Capacidade | O que significa |
|---|---|
| **Scaffolding** | Gera estruturas idiomáticas para o tipo de projeto escolhido |
| **Runtime** | Executa o projeto via `gaver run`, lendo comandos e configurações do `gaver.json` |
| **Módulos de domínio** | Cada módulo encapsula uma seção da regra de negócio, registrado explicitamente no engine |
| **Infraestrutura IA** | Gera contexto, instruções e memória persistente para agentes operarem no projeto |

---

## Princípios fundamentais

**Modular por domínio.** Os módulos do projeto gerado seguem a regra de negócio — cada módulo é autossuficiente, pode interagir com múltiplas tabelas e é registrado explicitamente no engine.

**Runtime próprio.** O projeto roda através do próprio Gaver. O `gaver.json` centraliza versão, banco de dados e comandos de inicialização — resolvendo variações entre plataformas (ex: `npm install` vs `pnpm install`).

**IA com escopo definido.** A infraestrutura de agentes gerada pelo Gaver delimita exatamente onde e como o agente pode atuar, gerencia suas limitações de contexto e persiste o histórico de implementações de forma cronológica.

**Consciente de plataforma, não preso a ela.** Cada tipo de projeto é um plugin independente. O core permanece agnóstico.

---

## Instalação

```sh
go install github.com/Dalistor/gaver/cmd/gaver@latest
```

---

## Comandos

### `gaver new`

Cria um novo projeto a partir de um template.

```sh
gaver new --type <tipo> --name <nome> [--database <banco>]
```

| Flag | Obrigatório | Descrição |
|---|---|---|
| `--type`, `-t` | Sim | Tipo do projeto: `api`, `webapp`, `agent`, `cli` |
| `--name`, `-n` | Sim | Nome do projeto |
| `--database`, `-d` | Não | Banco de dados: `postgres`, `mysql`, `sqlite` |

**Exemplos:**

```sh
gaver new --type api --name meu-servico
gaver new --type api --name meu-servico --database postgres
gaver new --type webapp --name meu-site --database sqlite
```

---

### `gaver init`

Inicializa as dependências do projeto. Lê o campo `commands.init` do `gaver.json` e o executa.

```sh
gaver init
```

---

### `gaver run`

Executa o projeto. Lê o campo `commands.run` do `gaver.json` e o executa.

```sh
gaver run
```

---

### `gaver build`

Compila o projeto. Lê o campo `commands.build` do `gaver.json` e o executa.

```sh
gaver build
```

---

### `gaver gen`

Gera artefatos para um projeto existente.

#### `gaver gen module`

Gera a estrutura de um novo módulo de domínio dentro do projeto atual. Deve ser executado na raiz do projeto (onde está o `gaver.json`).

```sh
gaver gen module --name <nome>
```

| Flag | Obrigatório | Descrição |
|---|---|---|
| `--name`, `-n` | Sim | Nome do módulo |

**Exemplos:**

```sh
gaver gen module --name orders
gaver gen module --name users
```

Após gerar, registre o módulo em `main.go`:

```go
e.Register(orders.New())
```


---

## Infraestrutura de agentes IA

Gaver gera, dentro de cada projeto, uma infraestrutura completa para que agentes IA trabalhem com contexto preciso e sem alucinação.

Os arquivos `_` em `ai/.pipeline/` são gerados automaticamente e **não devem ser editados**:

| Arquivo | Conteúdo |
|---|---|
| `_context.md` | Escopo do projeto: o que é, o que não é, onde o agente pode atuar |
| `_instructions.md` | Como o agente deve operar: fluxo, regras de gerenciamento de contexto |
| `_structure.md` | Mapa do projeto: arquivos, módulos e responsabilidades |

Para regras customizadas, adicione arquivos `.md` em `ai/.pipeline/custom/`. O agente lê os `_` arquivos como base e complementa com tudo em `custom/`.

Ao final de cada implementação, um agente verificador documenta as mudanças em `ai/.memory/implements/XXXX.md` e atualiza o resumo em `ai/.memory/_status.md`.

---

## Tipos de projeto

| Tipo | Documentação |
|---|---|
| `api` | [docs/api.md](docs/api.md) |

---

## Plataformas-alvo

- **Go** — serviços, CLIs, bibliotecas
- **Kotlin** — apps Android, backends JVM, multiplatform
- **Python** — scripts, pipelines de ML, frameworks de agentes
- **TypeScript / Node.js** — APIs, ferramentas de frontend, CLIs

---

## Autor

Feito por [Dalistor](https://github.com/Dalistor).

## Licença

Distribuído sob a [Licença MIT](LICENSE). Livre para usar, modificar e distribuir — inclusive em projetos comerciais.
