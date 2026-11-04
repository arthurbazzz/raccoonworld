# RW-007 — Plano técnico: Deploy em produção com Docker Compose

Status: **Aprovado** (2026-11-04) · Spec: [spec.md](spec.md) · Criado: 2026-11-04

## 1. Visão geral

```
                         VPS (porta 80)
Navegador ──http/ws──► traefik ─┬─ /ws, /maps/*, /health ──► server (Go, :8080) ──► postgres (:5432)
                                └─ /* ─────────────────────► web (Next standalone, :3000)

GitHub Actions: ci (lint + testes + smoke) ──► publish (GHCR) ──► deploy (SSH na VPS)
```

Arquivos novos:

```
server/Dockerfile
server/.dockerignore
web/Dockerfile
web/.dockerignore
deploy/compose.yaml           # stack de produção (imagens do GHCR)
deploy/compose.build.yaml     # override local: build a partir do código (RN10)
deploy/.env.example
deploy/README.md              # preparo da VPS (RN15)
.github/workflows/ci.yml      # PR e push: lint, testes, smoke
.github/workflows/deploy.yml  # push na main: publish + deploy
server/features/deploy/       # smoke da stack (godog), seção 5
specs/007-deploy-docker/features/deploy.feature
```

Mudanças em código existente: só `web/next.config.ts` (`output: "standalone"`) e `.gitignore`
(`deploy/.env`). Nenhuma dependência Go ou npm nova.

## 2. Imagens (RN1–RN3)

### 2.1 server

```dockerfile
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
```

- O mapa da Praça é lido em runtime? Se `server` carrega arquivos do disco (ex.: mapas), eles
  são copiados para a imagem; a task T2 confere isso antes de fechar o Dockerfile.
- O healthcheck não pode usar `curl` (distroless não tem shell). Solução: um subcomando
  `/server healthcheck` que faz `GET http://127.0.0.1:8080/health` e sai com 0/1. Fica em
  `cmd/server` (camada infra), sem tocar domínio nem casos de uso.

### 2.2 web

```dockerfile
FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
ARG NEXT_PUBLIC_WS_URL
ENV NEXT_PUBLIC_WS_URL=$NEXT_PUBLIC_WS_URL
RUN npm run build              # roda "sprites" antes (prebuild)

FROM node:22-alpine
WORKDIR /app
ENV NODE_ENV=production PORT=3000 HOSTNAME=0.0.0.0
COPY --from=build --chown=node /app/.next/standalone ./
COPY --from=build --chown=node /app/.next/static ./.next/static
COPY --from=build --chown=node /app/public ./public
USER node
EXPOSE 3000
CMD ["node", "server.js"]
```

- Healthcheck: `wget -qO- http://127.0.0.1:3000/` (alpine tem `wget`).
- `.dockerignore` exclui `node_modules`, `.next`, `public/sprites` (gerado no build) e `e2e`.

### 2.3 Tags (RN3)

`ghcr.io/<owner>/raccoonworld-server:{sha,latest}` e `...-web:{sha,latest}`. O compose de
produção usa `${IMAGE_TAG:-latest}`; o deploy grava o SHA em `IMAGE_TAG` (seção 4.2).

## 3. Stack — `deploy/compose.yaml` (RN4–RN10)

```yaml
name: raccoonworld

services:
  traefik:
    image: traefik:v3
    command:
      - --providers.docker=true
      - --providers.docker.exposedbydefault=false
      - --entrypoints.web.address=:80
      - --api.dashboard=false
    ports: ["80:80"]
    volumes: ["/var/run/docker.sock:/var/run/docker.sock:ro"]
    restart: unless-stopped

  server:
    image: ghcr.io/${GHCR_OWNER}/raccoonworld-server:${IMAGE_TAG:-latest}
    environment:
      RW_ADDR: ":8080"
      RW_ALLOWED_ORIGINS: ${PUBLIC_HOST}
      DATABASE_URL: postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable
    depends_on:
      postgres: { condition: service_healthy }
    healthcheck:
      test: ["CMD", "/server", "healthcheck"]
      interval: 10s
    stop_grace_period: 30s
    restart: unless-stopped
    labels:
      - traefik.enable=true
      - traefik.http.routers.server.rule=PathPrefix(`/ws`) || PathPrefix(`/maps/`) || Path(`/health`)
      - traefik.http.routers.server.priority=100
      - traefik.http.services.server.loadbalancer.server.port=8080

  web:
    image: ghcr.io/${GHCR_OWNER}/raccoonworld-web:${IMAGE_TAG:-latest}
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:3000/"]
      interval: 10s
    restart: unless-stopped
    labels:
      - traefik.enable=true
      - traefik.http.routers.web.rule=PathPrefix(`/`)
      - traefik.http.routers.web.priority=1
      - traefik.http.services.web.loadbalancer.server.port=3000

  postgres:
    image: postgres:17-alpine
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    volumes: ["pgdata:/var/lib/postgresql/data"]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $${POSTGRES_USER} -d $${POSTGRES_DB}"]
      interval: 5s
    restart: unless-stopped

volumes:
  pgdata:
```

- Nenhum serviço além do Traefik tem `ports` (RN5, RN6). Todos ficam na rede padrão do compose.
- `RW_DEBUG`, `RW_CUPS_FAST`, `RW_RUMMY_FAST` e `RW_RUMMY_SEED` não aparecem: ficam desligadas
  (RN9).
- **`NEXT_PUBLIC_WS_URL`** é `ws://${PUBLIC_HOST}/ws`, passado como build-arg no CI.
  Consequência: a imagem `web` fica presa ao IP. Trocar de IP exige novo build (aceitável para
  teste; com domínio isso deixa de importar).
- **Acesso ao docker.sock pelo Traefik**: montado só leitura. Risco conhecido e aceito para
  este cenário de teste.

### 3.1 `.env.example`

```
GHCR_OWNER=arthurbazzz
PUBLIC_HOST=203.0.113.10
IMAGE_TAG=latest
POSTGRES_USER=raccoon
POSTGRES_PASSWORD=troque-me
POSTGRES_DB=raccoonworld
```

### 3.2 Stack local (RN10)

`deploy/compose.build.yaml` sobrescreve `image` por `build` (contexto `../server` e `../web`,
com `NEXT_PUBLIC_WS_URL=ws://localhost/ws`). Uso:

```
docker compose -f deploy/compose.yaml -f deploy/compose.build.yaml --env-file deploy/.env.example up --build -d
```

com `PUBLIC_HOST=localhost`.

## 4. Pipeline (RN11–RN14, Q5)

### 4.1 `ci.yml` — em `pull_request` e `push` na `main`

| Job | Passos |
|---|---|
| `server` | `setup-go` (cache) → golangci-lint → `go test ./...` |
| `web` | `setup-node` (cache npm) → `npm ci` → lint → `format:check` → `npm test` |
| `smoke` | depende dos dois → sobe a stack local (3.2) → smoke godog (seção 5) → `down -v` |

Sem `test:e2e` (Q5).

### 4.2 `deploy.yml` — em `push` na `main`

Chamado depois do `ci.yml` (via `workflow_run` com `conclusion == success`), para nada ser
publicado com teste vermelho (RN11).

1. **publish**: `docker/login-action` no GHCR com o `GITHUB_TOKEN` (`packages: write`);
   `docker/build-push-action` para as duas imagens, tags `sha` e `latest`, cache `type=gha`.
   O build-arg `NEXT_PUBLIC_WS_URL` vem do secret `VPS_HOST`.
2. **deploy**: via `ssh` (chave em secret, `known_hosts` em secret para não aceitar qualquer
   host):
   - copia `deploy/compose.yaml` para `~/raccoonworld/` com `scp`;
   - grava `~/raccoonworld/.env` a partir do secret `PROD_ENV` + `IMAGE_TAG=<sha>`;
   - `docker compose pull && docker compose up -d --wait`;
3. **verifica** (RN14): `curl --fail --retry 10 http://$VPS_HOST/health`.

`up -d --wait` só termina quando os healthchecks passam. Se a nova versão não fica saudável,
o passo falha. Limite honesto de RN14: o compose não faz rollback sozinho; com um único
`server` há uma janela curta de indisponibilidade no `up`, e numa falha o workflow roda
`IMAGE_TAG=<tag anterior> docker compose up -d` (a tag anterior é lida do `.env` antes de
sobrescrever).

Secrets: `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`, `VPS_KNOWN_HOSTS`, `PROD_ENV`. Nenhum é
impresso: o `.env` é escrito com `printf '%s' "$PROD_ENV"` e sem `set -x` (RN13).

Na VPS, o `docker login ghcr.io` é feito uma vez à mão com um PAT só de `read:packages`
(RN15), já que as imagens são privadas.

## 5. Smoke da stack (Q6)

Para seguir a regra de BDD do projeto (godog no backend), o smoke é uma **terceira suíte
godog**, `TestDeployStack`, em `server/features/deploy/`, lendo
`specs/007-deploy-docker/features/deploy.feature`.

- Só roda com `RW_STACK_URL` definido (ex.: `http://localhost`); sem ela, `t.Skip`. Assim o
  `go test ./...` normal não depende de Docker.
- Os steps usam só a biblioteca padrão, `coder/websocket` (já no `go.mod`) e o CLI `docker`:

| Cenário | Como confere |
|---|---|
| Página pelo IP | `GET /` → 200 e HTML do Next |
| WebSocket pelo proxy | abre `ws://<host>/ws` com `Origin` permitido, faz o join e recebe o estado da Praça |
| Mapa pelo proxy | `GET /maps/praca` → 200 e JSON do mapa |
| Só a porta 80 | `docker compose ps --format json`: só o `traefik` publica porta, e é a 80 |
| Postgres persiste | `docker compose exec postgres psql` cria `probe`, `restart` da stack, lê de novo |
| Origem desconhecida | `ws://<host>/ws` com `Origin: http://evil.example` → recusado (403) |
| Push atualiza a VPS | coberto pelo passo 3 do `deploy.yml` (`/health` + tag), não pela suíte |

O cenário de persistência usa `docker compose restart`/`stop`+`up`, nunca `down -v`.

## 6. Riscos e pontos a confirmar

- **Arquivos lidos em runtime pelo server** (mapas, `testdata`): conferir na T2.
- **`next build` com `output: "standalone"`** pode exigir ajuste no carregamento dos sprites
  de `public/`; conferido na T3.
- **Cota do Actions**: estimativa de ~8 min por push na `main` (testes + smoke + build com
  cache). Folgado para 2.000 min/mês.
- **Sem TLS**: tráfego, inclusive chat, vai em texto puro. Aceito para teste (Q2).
