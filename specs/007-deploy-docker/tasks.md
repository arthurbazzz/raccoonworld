# RW-007 — Tasks: Deploy em produção com Docker Compose

Status: **Em andamento** · Aprovado: 2026-11-04 · Plano: [plan.md](plan.md) · Criado: 2026-11-04

Regras: uma task por vez, teste antes do código, commits `RW-007/Tn: <descrição>`. A cada task,
as suítes existentes continuam verdes (godog, `go test`, Vitest, lint).

## Fase 0 — Cenários

- [x] **T1** Passar os 6 cenários de stack da spec para
  `specs/007-deploy-docker/features/deploy.feature` (o 7º, "Push na main", fica como comentário:
  é verificado pelo `deploy.yml`). Criar `server/features/deploy/` com `TestDeployStack`, que dá
  `t.Skip` sem `RW_STACK_URL` e roda em modo estrito com ela. Conferir: sem a variável, o
  `go test ./...` segue verde; com ela, os 6 cenários aparecem **indefinidos e falhando**.

## Fase 1 — Imagens

- [ ] **T2** Subcomando `server healthcheck` em `cmd/server` (teste: sai 0 com `/health` 200 e
  1 caso contrário). Conferir se o `server` lê arquivos do disco em runtime (mapas etc.) e, se
  lê, incluí-los na imagem (plano, seção 6).
- [ ] **T3** `server/Dockerfile` e `server/.dockerignore` (plano, 2.1). Conferir: `docker build`
  passa, a imagem roda como não-root e `/health` responde.
- [ ] **T4** `output: "standalone"` em `web/next.config.ts`; `web/Dockerfile` e
  `web/.dockerignore` com o build-arg `NEXT_PUBLIC_WS_URL` (plano, 2.2). Conferir: a imagem
  serve a Praça com os sprites e roda como não-root; `npm run dev` e os testes seguem iguais.

## Fase 2 — Stack

- [ ] **T5** `deploy/compose.yaml`, `deploy/compose.build.yaml`, `deploy/.env.example` e
  `deploy/.env` no `.gitignore` (plano, seção 3). Conferir à mão: a stack local sobe com
  `--wait` e o jogo abre em `http://localhost`.
- [ ] **T6** Steps do smoke: página, WebSocket, mapa e origem recusada (plano, seção 5).
  Contra a stack local, esses 4 cenários ficam verdes.
- [ ] **T7** Steps do smoke: só a porta 80 e persistência do Postgres. Os 6 cenários ficam
  verdes.

## Fase 3 — Pipeline

- [ ] **T8** `.github/workflows/ci.yml`: jobs `server`, `web` e `smoke` (plano, 4.1). Conferir
  num PR que os três passam e que um teste quebrado de propósito deixa o PR vermelho.
- [ ] **T9** `.github/workflows/deploy.yml`, parte **publish**: build e push das duas imagens no
  GHCR com tags `sha` e `latest`, só depois do CI verde (plano, 4.2). Conferir os pacotes
  privados no GHCR.
- [ ] **T10** `deploy/README.md`: preparo da VPS (Docker, usuário de deploy, firewall na 80,
  login no GHCR com PAT `read:packages`) e a lista de secrets (RN15). **Feito junto com o dono**,
  que prepara a VPS e cadastra os secrets seguindo o guia.
- [ ] **T11** `deploy.yml`, parte **deploy**: SSH, cópia do compose, `.env` a partir do secret,
  `pull` + `up -d --wait`, checagem do `/health` e volta para a tag anterior em caso de falha
  (plano, 4.2). Conferir: um push na `main` atualiza a VPS e o jogo abre em `http://<IP>`; um
  deploy com healthcheck quebrado de propósito volta para a versão anterior.

## Fase 4 — Fechamento

- [ ] **T12** Atualizar `CLAUDE.md` (comandos de deploy e do smoke) e o `README.md`; marcar a
  spec como implementada.
