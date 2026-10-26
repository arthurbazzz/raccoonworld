# RW-005 — Tasks: Minigame dos três copos

Status: **Em andamento** · Aprovado: 2026-10-26 · Plano: [plan.md](plan.md) · Criado: 2026-10-26

Regras: uma task por vez, teste antes do código, commits `RW-005/Tn: <descrição>`. A cada task,
as suítes existentes continuam verdes (godog, `go test`, Vitest e os 14 cenários de navegador).

## Fase 0 — Cenários

- [ ] **T1** Passar os 22 cenários da spec para `specs/005-tres-copos/features/tres_copos.feature`,
  com as tags `@smoke` e `@e2e` do plano (seção 5). Registrar os steps no godog como pendentes:
  a suíte roda e mostra os 22 cenários **falhando/pendentes**.

## Fase 1 — Domínio (Go, TDD com `go test`)

- [ ] **T2** Barraca no mapa: tile `B` sólido e o objeto `booth` no `praca.json`; `TileMap.Booth()`
  e `TileMap.NearBooth(pos)` com alcance de 48 px (RC1, RC2). Tabela de casos: dentro, na borda
  e fora do alcance, nos 4 lados.
- [ ] **T3** `domain/cups`: tipos, constantes (seção 3.1), slots e `Shuffler`. Função pura que
  aplica as trocas e devolve onde a noz termina (RC17).
- [ ] **T4** `Game`: Idle → Lobby → Round (Show → Shuffle → Choose → Reveal) × 5 → Results →
  Lobby/Idle, avançado por `Advance(now)` e devolvendo eventos (RC4, RC5, RC7, RC8, RC11).
- [ ] **T5** `Pick` e `Leave`: uma escolha por rodada, só na fase Choose e só de quem joga;
  revelar cedo quando todos escolheram ou quando quem faltava sai; Idle quando todos saem;
  placar ordenado com empate (RC9, RC11–RC15, RC18).
- [ ] **T6** `Room.SetBusy`: quem joga não anda, e o input é zerado ao sair (RC21).

## Fase 2 — Casos de uso e protocolo (Go)

- [ ] **T7** `JoinCups`, `PickCup` e `LeaveCups`; o `RoomRunner` chama `Advance` em todo tick e
  traduz os eventos; `LeaveRoom` e a expiração tiram o jogador da partida. Steps godog com
  `FakeClock` e `Shuffler` fixo: **os 22 cenários passam** no godog.
- [ ] **T8** Protocolo v1: mensagens `cups.*` no `adapter/ws`, `cups` no `room.joined` e testes
  que fixam o JSON exato de cada mensagem. Smoke `@smoke` com o servidor real.
- [ ] **T9** `RW_CUPS_FAST=1` na config, dividindo todos os tempos por 10 (padrão: desligado).

## Fase 3 — Frontend (Vitest antes, navegador depois)

- [ ] **T10** Protocolo TS das mensagens `cups.*` (mesmos JSONs dos testes Go), `nearBooth` e o
  reducer `cupsState`.
- [ ] **T11** `timeline.ts`: fase e posição de cada copo pelo tempo decorrido, com o arco nas
  trocas; mesma resposta para qualquer `t`, inclusive depois de pular tempo (aba em segundo plano).
- [ ] **T12** Na Praça (PixiJS): barraca e copos com `Graphics` e profundidade, copos seguindo a
  `timeline`, aviso "Aperte E para jogar", tecla E, copinho ao lado do nome e
  `data-cups` na lista acessível (RC2, RC20, RC22).
- [ ] **T13** `CupsPanel.tsx`: espera com inscritos e contagem, copos como botões, teclas 1/2/3 e
  Esc (fora do chat), resultado da rodada, placar, **Jogar de novo** e **Sair** (RC6, RC8, RC12,
  RC13, RC19).
- [ ] **T14** Steps do playwright-bdd para os cenários `@e2e`, lendo `cups.round` pelo
  WebSocket para achar o copo certo, com o servidor em `RW_CUPS_FAST=1`.

## Fase 4 — Fechamento

- [ ] **T15** Rodar todas as suítes, lint e formatação; atualizar README, CLAUDE.md e a seção
  "O que mudou durante a implementação" do plano; pedir ao dono para jogar uma partida com
  duas abas e aprovar.

## Definição de Pronto

- [ ] Os 22 cenários passam no godog; os `@smoke` com o servidor real; os `@e2e` no navegador
- [ ] Os 14 cenários de navegador das RW-001/RW-002 continuam passando
- [ ] `go test`, Vitest, `golangci-lint` (com as regras de camada), `eslint` e Prettier limpos
- [ ] spec, plan e tasks atualizados com o que foi feito
- [ ] Partida jogada e aprovada pelo dono
