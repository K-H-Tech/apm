# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

APM (Assistant Product Manager) is an AI-powered product management assistant built in Go. It helps product managers with PRD generation, Jira/Confluence integration, backlog prioritization, and product discovery workflows.

**Type**: Standalone HTTP service
**Auth**: OAuth 2.0 with Atlassian
**LLM**: Multi-provider (OpenAI + Anthropic)

## Build & Test Commands

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./pkg/priority/...

# Run a single test
go test -run TestRICECalculator ./pkg/priority/

# Build the server
go build -o bin/apm-server ./cmd/server

# Run the server
./bin/apm-server -c config.yaml

# Format and vet code
go fmt ./... && go vet ./...

# Run with Docker Compose (local dev)
docker-compose -f docker/docker-compose.yaml up -d
```

## Architecture

### Directory Structure

```
apm/
├── cmd/server/              # Application entry point
├── internal/
│   ├── config/              # Viper config with TB_ env prefix
│   ├── contract/            # Service interfaces (PRD, LLM, Jira, Priority)
│   ├── models/              # Domain models (PRD, Template, Priority, User)
│   ├── repository/          # Data access layer (PostgreSQL)
│   ├── service/             # Business logic
│   ├── handlers/            # Gin HTTP handlers
│   └── middleware/          # Auth, rate limiting, metrics
├── pkg/
│   ├── llm/                 # LLM providers (OpenAI, Anthropic)
│   │   └── prompts/         # Prompt templates for AI generation
│   ├── priority/            # RICE, MoSCoW, ICE, Weighted calculators
│   ├── atlassian/           # Jira/Confluence clients + OAuth
│   ├── crypt/               # AES-GCM encryption
│   └── jwt/                 # JWT auth
├── tools/                   # Database connection helpers
├── migrations/              # SQL migrations
└── docker/                  # Docker & Compose files
```

### Key Patterns

**Contracts**: All major components defined as interfaces in `internal/contract/`. Implementations in `pkg/` or `internal/service/`.

**Factory Methods**: Use `New*` constructors (e.g., `NewPRDService`, `NewOpenAIProvider`, `NewRICECalculator`).

**Repository Pattern**: Data access through repository interfaces, implementations use PostgreSQL.

**Configuration**: Viper with YAML + environment variables (`TB_` prefix).

### Core Domain Models

**PRD** (`internal/models/prd.go`):
- Structured content stored as JSONB
- Status workflow: draft → review → approved → published
- Links to Jira epics and Confluence pages

**Priority Score** (`internal/models/priority.go`):
- Multi-framework support per PRD
- RICE, MoSCoW, ICE, Weighted scoring

**Template** (`internal/models/template.go`):
- Reusable PRD structures
- Types: feature, bug_fix, enhancement, technical

### API Structure

```
/auth/atlassian           # OAuth 2.0 flow
/api/v1/prds              # PRD CRUD + AI generation
/api/v1/templates         # Template management
/api/v1/priorities        # Scoring frameworks
/api/v1/jira              # Jira integration
/api/v1/confluence        # Confluence integration
/api/v1/ai                # AI generation endpoints
```

### Dependencies

- Web: Gin
- Database: PostgreSQL (lib/pq), Redis (go-redis/v9)
- Config: Viper
- Metrics: Prometheus
- JWT: golang-jwt/v5
- LLM: OpenAI API, Anthropic API

## Custom Slash Commands

Available agent personas in `.claude/commands/`:

| Command | Description |
|---------|-------------|
| `/senior-golang-engineer` | Go expert for code review, architecture, concurrency, and performance |
| `/senior-product-manager` | Product strategy, prioritization frameworks, requirements |
| `/senior-ai-engineer` | LLM integration, prompt engineering, AI system design |
| `/senior-data-analyst` | SQL, metrics design, data analysis |

## APM-Specific Guidelines

### When Working on PRD Features
- PRD content is structured JSONB - see `internal/models/prd.go` for schema
- Always validate against template schema when creating/updating
- Version all content changes in `prd_versions` table

### When Working on LLM Integration
- Use `pkg/llm/provider.go` interface for all LLM calls
- Add prompts to `pkg/llm/prompts/` with clear documentation
- Handle rate limiting and token usage tracking

### When Working on Atlassian Integration
- OAuth tokens stored encrypted (use `pkg/crypt`)
- All Jira/Confluence calls go through `pkg/atlassian/` clients
- Track sync status in `jira_sync_log` / `confluence_sync_log` tables

### When Working on Priority Frameworks
- Calculators in `pkg/priority/` are stateless
- Store scores via `internal/repository/priority_repository.go`
- Support multiple frameworks per PRD
