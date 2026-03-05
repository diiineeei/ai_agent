# AI Agent

Agente de IA conversacional com suporte a múltiplos agentes configuráveis, upload de documentos e busca semântica.

## Stack

- **Backend:** Go 1.25 · Google Gemini API · MongoDB
- **Frontend:** Vue 3 · Vuetify 3 · Pinia · Vite

## Funcionalidades

- Chat com histórico persistente por sessão
- Múltiplos agentes configuráveis (modelo, instrução de sistema, skills)
- Upload de arquivos `.txt` e `.pdf` diretamente pelo chat
- Busca semântica nos documentos enviados
- Consulta de clima em tempo real
- Interface responsiva com suporte a tema claro/escuro

## Pré-requisitos

- Go 1.25+
- Node.js 18+
- Docker e Docker Compose
- Chave de API do [Google Gemini](https://aistudio.google.com/apikey)

## Configuração

Crie um arquivo `.env` na raiz (ou exporte as variáveis):

```env
GEMINI_API_KEY=sua_chave_aqui
MODEL=gemini-2.5-flash        # opcional
HTTP_PORT=8080                # opcional
MONGODB_URI=mongodb://localhost:27017  # opcional
MONGODB_DB=ai_agent           # opcional
```

## Rodando o projeto

```bash
# 1. Subir o MongoDB
make up

# 2. Iniciar o servidor Go
make run

# 3. Em outro terminal, iniciar o frontend
make web
```

O frontend estará disponível em `http://localhost:5173` e o backend em `http://localhost:8080`.

## API

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/prompt` | Enviar mensagem ao agente |
| `GET` | `/sessions` | Listar sessões |
| `GET` | `/history` | Histórico de uma sessão |
| `DELETE` | `/history` | Apagar histórico |
| `GET/POST` | `/agent-configs` | Listar / criar agentes |
| `GET/PUT/DELETE` | `/agent-configs/{id}` | Buscar / editar / excluir agente |
| `GET` | `/files` | Listar arquivos enviados |
| `POST` | `/files` | Enviar arquivo |
| `DELETE` | `/files/{id}` | Remover arquivo |
| `GET` | `/skills` | Listar skills |
| `PUT` | `/skills/{name}/toggle` | Ativar / desativar skill |

### Exemplo de uso

```bash
# Criar um agente
curl -X POST http://localhost:8080/agent-configs \
  -H 'Content-Type: application/json' \
  -d '{"name":"Assistente","model":"gemini-2.5-flash","system_instruction":"Responda sempre em português.","enabled_skills":["weather","search_documents"]}'

# Enviar uma mensagem (nova sessão)
curl -X POST http://localhost:8080/prompt \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"s1","agent_config_id":"<id>","prompt":"Olá!"}'

# Continuar a mesma sessão (sem agent_config_id)
curl -X POST http://localhost:8080/prompt \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"s1","prompt":"Como está o tempo em São Paulo?"}'
```

## Estrutura do projeto

```
.
├── cmd/server/         # Entrypoint e wiring de dependências
├── internal/
│   ├── agent/          # Cliente Gemini e embeddings FNV
│   ├── handler/        # HTTP handlers
│   ├── model/          # Structs e tipos
│   ├── repository/     # Acesso ao MongoDB
│   └── skills/         # Registry e implementação das skills
├── web/                # Frontend Vue 3
│   └── src/
│       ├── views/      # ChatView, AgentsView, FilesView, SkillsView
│       ├── stores/     # Pinia stores
│       └── services/   # Chamadas à API
├── docker-compose.yml
└── Makefile
```

## Makefile

| Comando | Descrição |
|---------|-----------|
| `make up` | Sobe o MongoDB via Docker |
| `make down` | Para os containers |
| `make run` | Inicia o servidor Go |
| `make web` | Instala dependências e inicia o frontend |
| `make fmt` | Formata o código Go |
| `make vet` | Roda o `go vet` |
| `make tidy` | Atualiza o `go.mod` |
