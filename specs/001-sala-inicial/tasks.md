# RW-001 â€” Tasks: Sala inicial

Status: **Aprovado** (2026-10-25) Â· Plano: [plan.md](plan.md) Â· Criado: 2026-10-25

Regras: execute uma task por vez, na ordem. Cada task de implementaÃ§Ã£o sÃ³ termina quando
os cenÃ¡rios ou testes indicados passam. Commits no formato `RW-001/Tn: <descriÃ§Ã£o>`.
`[P]` indica uma task que pode rodar em paralelo com a anterior.

## Fase 0 â€” FundaÃ§Ã£o

- [x] **T1** Criar o esqueleto do monorepo: `server/` (`go mod init`), `web/` (Next.js + TS),
  README com os comandos. Pronto quando: `go build ./...` e `npm run build` passam vazios.
- [x] **T2** Configurar lint e formataÃ§Ã£o: `golangci-lint` no server, `eslint` e `prettier` no web.
- [x] **T3** Escrever os `.feature` em `specs/001-sala-inicial/features/`, a partir dos 14 cenÃ¡rios
  da spec: `entrar.feature`, `movimento.feature`, `chat.feature`, `conexao.feature`.
- [x] **T4** Montar o runner godog em `server/features/` lendo esses `.feature`, com
  `FakeClock` e `FakeBroadcaster` e todos os steps pendentes.
  O godog roda com o filtro de tags `~@ui`.
  Pronto quando: `go test ./features/...` roda e mostra todos os cenÃ¡rios nÃ£o-`@ui`
  **falhando** ou pendentes.

## Fase 1 â€” DomÃ­nio (go test, TDD)

- [x] **T5** `Nickname`: value object com validaÃ§Ã£o (RN1). Testes de tabela.
- [x] **T6** `Vec` e `Input`: vetor a partir de WASD, com a diagonal normalizada (RN3).
- [x] **T7** `TileMap`: carregar `maps/praca.json`, colisÃ£o eixo a eixo e limites da sala.
- [x] **T8** `Room`: entrar e sair, capacidade de 10, apelido Ãºnico e `Step(dt)` aplicando
  o movimento com colisÃ£o.
- [x] **T9** `Chat`: limite de 120 caracteres e intervalo mÃ­nimo de 1 s entre mensagens (RN4).

## Fase 2 â€” Casos de uso + BDD do backend

- [x] **T10** Portas `Broadcaster` e `Clock`, mais a goroutine da sala (channel de comandos + tick de 20 Hz).
- [x] **T11** `JoinRoom`: implementar os steps de `entrar.feature` (entrada, apelido em uso,
  apelido invÃ¡lido, sala cheia, aviso aos outros). Pronto quando esses cenÃ¡rios passam.
- [x] **T12** `SetInput`: steps de `movimento.feature` (andar, diagonal, parar, obstÃ¡culo, limite).
- [x] **T13** `SendChat`: steps de `chat.feature` (mensagem, mensagens rÃ¡pidas demais).
- [x] **T14** `LeaveRoom` + timeout: steps de `conexao.feature` (desconexÃ£o em atÃ© 10 s, com `FakeClock`).

## Fase 3 â€” Transporte

- [x] **T15** Protocolo v1: DTOs de envelope, encode/decode e testes de ida e volta.
- [x] **T16** Gin: `GET /health`, `GET /ws` com coder/websocket, checagem de origem e
  ping/pong (5 s / 10 s). ComposiÃ§Ã£o em `cmd/server/main.go`.
- [x] **T17** Smoke BDD com o servidor real: entrar e chat via WebSocket de verdade.

## Fase 4 â€” Frontend

- [x] **T18** [P] Cliente WS tipado em `web/src/net/` (tipos do protocolo v1, reconexÃ£o simples).
- [x] **T19** Tela de apelido, com as mensagens para cada `join_rejected`.
- [x] **T20** Render PixiJS: tilemap da PraÃ§a, sprite placeholder do guaxinim, ordenaÃ§Ã£o por `y` (RN6).
- [x] **T21** Input WASD (enviado sÃ³ quando muda) e bloqueio enquanto o chat tem foco (RN3.1).
- [ ] **T22** InterpolaÃ§Ã£o dos snapshots, com cerca de 100 ms de atraso.
- [ ] **T23** Chat: campo de texto e balÃ£o sobre o sprite por 5 s.

## Fase 5 â€” BDD de UI e fechamento

- [ ] **T24** playwright-bdd com dois navegadores: entrar, andar, chat e digitar sem mover.
- [ ] **T25** Checar a DefiniÃ§Ã£o de Pronto da constituiÃ§Ã£o e atualizar spec e plan com o que mudou.
