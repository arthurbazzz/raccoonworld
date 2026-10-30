# RW-006 — Tasks: Rummy das Nozes

Status: **Em andamento** · Aprovado: 2026-10-30 · Plano: [plan.md](plan.md) · Criado: 2026-10-30

Regras: uma task por vez, teste antes do código, commits `RW-006/Tn: <descrição>`. A cada task,
as suítes existentes continuam verdes (godog, `go test`, Vitest e os cenários de navegador da
RW-001 a RW-005).

## Fase 0 — Cenários

- [x] **T1** Passar os 20 cenários da spec para `specs/006-rummy/features/rummy.feature`, com
  `@smoke` e `@e2e` conforme o plano (seção 5). Ligar o arquivo nas suítes do godog: os 20
  aparecem **indefinidos e falhando** (modo estrito).

## Fase 1 — Domínio (Go, TDD com `go test`)

- [x] **T2** Mesa no mapa: tile `M` sólido nas colunas 13–16, linhas 15–16; `TileMap.NearTable`
  com alcance de 48 px; o `parseTileMap` do navegador aprende o `M`, com `nearTable` e a mesma
  tabela de casos (RN4, RN25).
- [x] **T3** `domain/rummy`: `Tile`, as 106 peças de `NewDeck` com IDs fixos, `Shuffler` e as
  constantes (plano, seção 3.1).
- [x] **T4** `Classify` e `Points`: grupos, sequências e coringas, com os motivos de erro. Tabela
  de casos grande, guardada em JSON em `server/internal/domain/rummy/testdata/melds.json` para o
  TypeScript reaproveitar (RN2, RN3).
- [x] **T5** `Match`: distribuição, ordem sorteada, rascunho (`Move` com a checagem de
  permutação), `Undo`, `Draw` e `EndTurn` com as regras de mesa válida e de nenhuma peça da
  mesa na mão (RN8–RN12, RN14).
- [x] **T6** Modos e abertura: Clássico com 30 pontos só da mão e sem mexer na mesa antes de
  abrir; Livre sem abertura (RN13).
- [x] **T7** Prazo e fim: vez de 40 s por `Advance`, monte vazio e volta de passes, vitória por
  mão vazia, placar com coringa valendo 30 (RN15–RN19).
- [x] **T8** Saída e ausência: peças de volta ao monte, vez passada se era dele, último jogador
  ganha; assento ausente guardado pelo apelido por 30 s e recuperado (RN20–RN22, plano 3.2).
- [x] **T9** `Table`: criar sala com modo, entrar, vagas, contagem de 15 s que para com 1,
  início na hora com 4, sala "jogando" fechada, "Jogar de novo" volta à espera, sala vazia some
  (RN5–RN7).

## Fase 2 — Casos de uso e protocolo (Go)

- [x] **T10** Casos de uso (`OpenTable`, `CreateTableRoom`, `JoinTableRoom`, `MoveTiles`,
  `EndTurn`, `DrawTile`, `UndoTurn`, `LeaveTableRoom`); o `RoomRunner` avança a mesa no tick;
  `busy` trava o guaxinim e impede jogar copos e rummy ao mesmo tempo; `JoinRoom` e `LeaveRoom`
  avisam a mesa. Steps godog com `FakeClock` e `Shuffler` fixo: **os 20 cenários passam**.
- [ ] **T11** Protocolo v1: mensagens `rummy.*` no `adapter/ws`, com `rummy.state` mandado a cada
  jogador separadamente e `rummy` no `room.joined`. Testes que fixam o JSON exato e que conferem
  que nenhuma mensagem leva a mão de outro jogador nem o monte. Smoke `@smoke` com servidor real.
- [ ] **T12** `RW_RUMMY_FAST=1` (espera de 1,5 s) e `RW_RUMMY_SEED=<n>` na config, desligados por
  padrão.

## Fase 3 — Arte

- [ ] **T13** `npm run sprites` recorta `art/rummy.webp`: 13 peças × 4 cores, coringa, verso e
  monte marrons, e a mesa, para `public/sprites/rummy.png` e `rummy.json`. Para com erro se a
  contagem não bater. Testes Vitest do recorte (Q6).

## Fase 4 — Frontend

- [ ] **T14** Lógica pura em `src/game/rummy/`: `tiles.ts` (ID → peça, mesmos dados do Go),
  `melds.ts` (lendo a tabela da T4), `rummyState.ts` (reducer) e o protocolo, com Vitest.
- [ ] **T15** Mesa na Praça: sprite na `scene.ts` ordenado por profundidade, aviso "Aperte E
  para jogar", número de salas jogando e etiqueta "jogando" sobre o guaxinim (RN25, RN27).
- [ ] **T16** `TableRoomsPanel`: lista de salas, criar com modo, entrar, espera com contagem
  (RN5–RN7).
- [ ] **T17** `RummyPanel`: mesa e mão com arrastar e soltar, destaque de inválidas, ordenar a mão,
  relógio, contagem de peças dos outros, ausentes, rascunho ao vivo de quem joga, botões e
  placar (RN9–RN19, RN26).

## Fase 5 — Navegador e fechamento

- [ ] **T18** Steps playwright-bdd dos cenários `@e2e`, com `RW_RUMMY_FAST` e `RW_RUMMY_SEED`
  para mãos conhecidas. **Todos os cenários verdes**, inclusive os das specs anteriores.
- [ ] **T19** Documentação: comandos novos no `CLAUDE.md`, seção "O que mudou durante a
  implementação" no plano e status final da spec, do plano e das tasks.
