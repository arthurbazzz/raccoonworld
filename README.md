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

O servidor sobe em `:8080`, com `GET /health` e `GET /ws` (WebSocket, protocolo v1).
Variáveis de ambiente:

| Variável | Padrão | Uso |
|---|---|---|
| `RW_ADDR` | `:8080` | endereço HTTP |
| `RW_ALLOWED_ORIGINS` | `localhost:3000` | origens do navegador aceitas no `/ws`, separadas por vírgula |
| `RW_DEBUG` | — | `1` liga os logs de debug |

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
npm test
```

```bash
npm run test:e2e
```

O `test:e2e` roda os cenários BDD de navegador (playwright-bdd) e sobe o servidor Go e o
Next.js sozinho. Na primeira vez, instale o Chromium do Playwright:

```bash
npx playwright install chromium
```

```bash
npm run lint
```

O endereço do servidor vem de `NEXT_PUBLIC_WS_URL` (padrão `ws://localhost:8080/ws`).

```bash
npm run format
```

**Arte do guaxinim (sprites):** a imagem fonte fica em `web/art/raccoon.png` (3 fileiras: frente,
costas e lado direito; 6 colunas: 2 parado + 4 andando; fundo transparente). `npm run sprites`
gera a folha alinhada em `web/public/sprites/` e roda sozinho antes de `dev`, `build` e
`test:e2e`. Para trocar a arte, basta substituir o arquivo.
