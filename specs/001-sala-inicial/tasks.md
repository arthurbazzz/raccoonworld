# RW-001 — Tasks: Sala inicial

Status: **Aprovado** (2026-10-25) · Plano: [plan.md](plan.md) · Criado: 2026-10-25

Regras: execute uma task por vez, na ordem. Cada task de implementação só termina quando
os cenários ou testes indicados passam. Commits no formato `RW-001/Tn: <descrição>`.
`[P]` indica uma task que pode rodar em paralelo com a anterior.

## Fase 0 — Fundação

- [x] **T1** Criar o esqueleto do monorepo: `server/` (`go mod init`), `web/` (Next.js + TS),
  README com os comandos. Pronto quando: `go build ./...` e `npm run build` passam vazios.
- [x] **T2** Configurar lint e formatação: `golangci-lint` no server, `eslint` e `prettier` no web.
- [x] **T3** Escrever os `.feature` em `specs/001-sala-inicial/features/`, a partir dos 14 cenários
  da spec: `entrar.feature`, `movimento.feature`, `chat.feature`, `conexao.feature`.
- [ ] **T4** Montar o runner godog em `server/features/` lendo esses `.feature`, com
  `FakeClock` e `FakeBroadcaster` e todos os steps pendentes.
  O godog roda com o filtro de tags `~@ui`.
  Pronto quando: `go test ./features/...` roda e mostra todos os cenários não-`@ui`
  **falhando** ou pendentes.

## Fase 1 — Domínio (go test, TDD)

- [ ] **T5** `Nickname`: value object com validação (RN1). Testes de tabela.
- [ ] **T6** `Vec` e `Input`: vetor a partir de WASD, com a diagonal normalizada (RN3).
- [ ] **T7** `TileMap`: carregar `maps/praca.json`, colisão eixo a eixo e limites da sala.
- [ ] **T8** `Room`: entrar e sair, capacidade de 10, apelido único e `Step(dt)` aplicando
  o movimento com colisão.
- [ ] **T9** `Chat`: limite de 120 caracteres e intervalo mínimo de 1 s entre mensagens (RN4).

## Fase 2 — Casos de uso + BDD do backend

- [ ] **T10** Portas `Broadcaster` e `Clock`, mais a goroutine da sala (channel de comandos + tick de 20 Hz).
- [ ] **T11** `JoinRoom`: implementar os steps de `entrar.feature` (entrada, apelido em uso,
  apelido inválido, sala cheia, aviso aos outros). Pronto quando esses cenários passam.
- [ ] **T12** `SetInput`: steps de `movimento.feature` (andar, diagonal, parar, obstáculo, limite).
- [ ] **T13** `SendChat`: steps de `chat.feature` (mensagem, mensagens rápidas demais).
- [ ] **T14** `LeaveRoom` + timeout: steps de `conexao.feature` (desconexão em até 10 s, com `FakeClock`).

## Fase 3 — Transporte

- [ ] **T15** Protocolo v1: DTOs de envelope, encode/decode e testes de ida e volta.
- [ ] **T16** Gin: `GET /health`, `GET /ws` com coder/websocket, checagem de origem e
  ping/pong (5 s / 10 s). Composição em `cmd/server/main.go`.
- [ ] **T17** Smoke BDD com o servidor real: entrar e chat via WebSocket de verdade.

## Fase 4 — Frontend

- [ ] **T18** [P] Cliente WS tipado em `web/src/net/` (tipos do protocolo v1, reconexão simples).
- [ ] **T19** Tela de apelido, com as mensagens para cada `join_rejected`.
- [ ] **T20** Render PixiJS: tilemap da Praça, sprite placeholder do guaxinim, ordenação por `y` (RN6).
- [ ] **T21** Input WASD (enviado só quando muda) e bloqueio enquanto o chat tem foco (RN3.1).
- [ ] **T22** Interpolação dos snapshots, com cerca de 100 ms de atraso.
- [ ] **T23** Chat: campo de texto e balão sobre o sprite por 5 s.

## Fase 5 — BDD de UI e fechamento

- [ ] **T24** playwright-bdd com dois navegadores: entrar, andar, chat e digitar sem mover.
- [ ] **T25** Checar a Definição de Pronto da constituição e atualizar spec e plan com o que mudou.
