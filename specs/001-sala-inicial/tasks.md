# RW-001 — Tasks: Sala inicial

Status: **Concluído** (2026-10-25) · Aprovado: 2026-10-25 · Plano: [plan.md](plan.md) · Criado: 2026-10-25

Regras: execute uma task por vez, na ordem. Cada task de implementação só termina quando
os cenários ou testes indicados passam. Commits no formato `RW-001/Tn: <descrição>`.
`[P]` indica uma task que pode rodar em paralelo com a anterior.

## Fase 0 — Fundação

- [x] **T1** Criar o esqueleto do monorepo: `server/` (`go mod init`), `web/` (Next.js + TS),
  README com os comandos. Pronto quando: `go build ./...` e `npm run build` passam vazios.
- [x] **T2** Configurar lint e formatação: `golangci-lint` no server, `eslint` e `prettier` no web.
- [x] **T3** Escrever os `.feature` em `specs/001-sala-inicial/features/`, a partir dos 14 cenários
  da spec: `entrar.feature`, `movimento.feature`, `chat.feature`, `conexao.feature`.
- [x] **T4** Montar o runner godog em `server/features/` lendo esses `.feature`, com
  `FakeClock` e `FakeBroadcaster` e todos os steps pendentes.
  O godog roda com o filtro de tags `~@ui`.
  Pronto quando: `go test ./features/...` roda e mostra todos os cenários não-`@ui`
  **falhando** ou pendentes.

## Fase 1 — Domínio (go test, TDD)

- [x] **T5** `Nickname`: value object com validação (RN1). Testes de tabela.
- [x] **T6** `Vec` e `Input`: vetor a partir de WASD, com a diagonal normalizada (RN3).
- [x] **T7** `TileMap`: carregar `maps/praca.json`, colisão eixo a eixo e limites da sala.
- [x] **T8** `Room`: entrar e sair, capacidade de 10, apelido único e `Step(dt)` aplicando
  o movimento com colisão.
- [x] **T9** `Chat`: limite de 120 caracteres e intervalo mínimo de 1 s entre mensagens (RN4).

## Fase 2 — Casos de uso + BDD do backend

- [x] **T10** Portas `Broadcaster` e `Clock`, mais a goroutine da sala (channel de comandos + tick de 20 Hz).
- [x] **T11** `JoinRoom`: implementar os steps de `entrar.feature` (entrada, apelido em uso,
  apelido inválido, sala cheia, aviso aos outros). Pronto quando esses cenários passam.
- [x] **T12** `SetInput`: steps de `movimento.feature` (andar, diagonal, parar, obstáculo, limite).
- [x] **T13** `SendChat`: steps de `chat.feature` (mensagem, mensagens rápidas demais).
- [x] **T14** `LeaveRoom` + timeout: steps de `conexao.feature` (desconexão em até 10 s, com `FakeClock`).

## Fase 3 — Transporte

- [x] **T15** Protocolo v1: DTOs de envelope, encode/decode e testes de ida e volta.
- [x] **T16** Gin: `GET /health`, `GET /ws` com coder/websocket, checagem de origem e
  ping/pong (5 s / 10 s). Composição em `cmd/server/main.go`.
- [x] **T17** Smoke BDD com o servidor real: entrar e chat via WebSocket de verdade.

## Fase 4 — Frontend

- [x] **T18** [P] Cliente WS tipado em `web/src/net/` (tipos do protocolo v1, reconexão simples).
- [x] **T19** Tela de apelido, com as mensagens para cada `join_rejected`.
- [x] **T20** Render PixiJS: tilemap da Praça, sprite placeholder do guaxinim, ordenação por `y` (RN6).
- [x] **T21** Input WASD (enviado só quando muda) e bloqueio enquanto o chat tem foco (RN3.1).
- [x] **T22** Interpolação dos snapshots, com cerca de 100 ms de atraso.
- [x] **T23** Chat: campo de texto e balão sobre o sprite por 5 s.

## Fase 5 — BDD de UI e fechamento

- [x] **T24** playwright-bdd com dois navegadores: entrar, andar, chat e digitar sem mover.
- [x] **T25** Checar a Definição de Pronto da constituição e atualizar spec e plan com o que mudou.

## Fechamento (T25, 2026-10-25)

### Definição de Pronto

- [x] **Critérios de aceitação da spec atendidos.** Todos os cenários de `features/` passam.
- [x] **Cenários BDD passando:**
  - godog, casos de uso: **24/24 cenários, 94/94 steps**;
  - godog, servidor real por WebSocket (`@smoke`): **3/3**;
  - playwright-bdd, navegador com dois jogadores (`@ui`/`@e2e`): **4/4**.
- [x] **Testes, lint e formatação limpos:**
  - `go test ./...`, `go vet`, `golangci-lint run` (0 problemas) e `golangci-lint fmt` sem diferenças;
  - Vitest com 77 testes, `tsc`, ESLint, Prettier e `next build`.
- [x] **Nenhuma violação de camada.** O depguard no `golangci-lint` barra imports proibidos no
  domínio e nos casos de uso, e passa limpo.
- [x] **spec, plan e tasks atualizados.** O plan §8 lista tudo o que mudou em relação ao aprovado.

Os testes de mutação feitos ao longo das tasks confirmaram que cada suíte pega o bug que
deveria pegar: diagonal, limite de chat, expiração, heartbeat, Hub e WASD no chat.

### Segurança das dependências

- `npm audit fix` atualizou o `sharp`. O `govulncheck` levou o `golang.org/x/net` para v0.60.0
  e o `quic-go` para v0.59.1.
- **Pendente:** o `postcss` embutido no Next 15 tem alertas, e a correção exige o **Next 16**
  (versão major). O risco é baixo, porque o postcss só processa o CSS do próprio projeto no build.
- **Pendente:** a biblioteca padrão do **Go 1.27.0** tem 7 alertas (`net/http`, `crypto/tls`,
  `crypto/x509`). A correção é instalar o Go 1.27.3 ou mais novo.

### Limitações conhecidas

- **Reconexão:** se a rede cai sem o servidor perceber, o guaxinim antigo fica até 10 s na sala,
  e a reconexão nesse intervalo recebe "apelido em uso". A próxima tentativa funciona.
- **Race detector (`go test -race`)** não foi rodado, porque precisa de cgo e de um compilador C.
- **Suavidade da interpolação** não foi conferida a olho. O painel do navegador usado nos testes
  renderizava a 2 quadros por segundo. A lógica está coberta pelos testes unitários.
- Arte provisória (formas simples) e uma sala só.

### Próximas specs sugeridas

- Chat restrito (filtro ou frases prontas), já previsto na spec.
- Autenticação: Firebase, Sign in with Apple e cadastro manual (constituição §6).
- Predição local do movimento, se o atraso incomodar (plan §7).
