# Raccoons World

Mundo virtual social com minigames. Desenvolvimento guiado por specs:
leia [`specs/constituion.md`](specs/constituion.md) e [`CLAUDE.md`](CLAUDE.md) antes de contribuir.

## Estrutura

```
server/   backend Go (Gin, WebSocket), Clean Architecture
web/      frontend Next.js + PixiJS
specs/    constituição, referências e specs (spec → plan → tasks)
```

## Requisitos

- Go 1.27+
- Node.js 24 LTS+

## Comandos

### Backend (`server/`)

```bash
go build ./...
```

```bash
go run ./cmd/server
```

```bash
go test ./...
```

```bash
go tool -modfile=tools/go.mod golangci-lint run ./...
```

```bash
go tool -modfile=tools/go.mod golangci-lint fmt ./...
```

O golangci-lint fica num módulo separado (`server/tools/`), para as dependências dele
não entrarem no `go.mod` do servidor.

### Frontend (`web/`)

```bash
npm install
```

```bash
npm run dev
```

```bash
npm run build
```

```bash
npm run lint
```

```bash
npm run format
```
