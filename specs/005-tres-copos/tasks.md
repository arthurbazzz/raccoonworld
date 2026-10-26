# RW-005 — Tasks: Minigame dos três copos

Status: **Em andamento** · Aprovado: 2026-10-26 · Plano: [plan.md](plan.md) · Criado: 2026-10-26

Regras: uma task por vez, teste antes do código, commits `RW-005/Tn: <descrição>`. A cada task,
as suítes existentes continuam verdes (godog, `go test`, Vitest e os 14 cenários de navegador).

## Fase 0 — Cenários

- [x] **T1** Passar os 22 cenários da spec para `specs/005-tres-copos/features/tres_copos.feature`,
  com as tags `@smoke` e `@e2e` do plano (seção 5). Registrar os steps no godog como pendentes:
  a suíte roda e mostra os 22 cenários **falhando/pendentes**.
  - Os cenários não precisaram de steps pendentes: sem definição, o godog já os mostra como
    indefinidos e falha (modo estrito). Os de navegador ficaram com `@e2e` e um com `@smoke @e2e`.

## Fase 1 — Domínio (Go, TDD com `go test`)

- [x] **T2** Barraca no mapa: tile `B` sólido e o objeto `booth` no `praca.json`; `TileMap.Booth()`
  e `TileMap.NearBooth(pos)` com alcance de 48 px (RC1, RC2). Tabela de casos: dentro, na borda
  e fora do alcance, nos 4 lados.
  - **Mudou:** a barraca vem só dos tiles `B` (colunas 14–16 da linha 4, caixa 448,128 96×32),
    sem um objeto `booth` separado no JSON. Uma fonte só, sem risco de as duas discordarem.
  - O `parseTileMap` do navegador também aprendeu o `B` (senão a Praça nem carregaria), com o
    `nearBooth` e a mesma tabela de casos do Go.
- [x] **T3** `domain/cups`: tipos, constantes (seção 3.1), slots e `Shuffler`. Função pura que
  aplica as trocas e devolve onde a noz termina (RC17).
- [x] **T4** `Game`: Idle → Lobby → Round (Show → Shuffle → Choose → Reveal) × 5 → Results →
  Lobby/Idle, avançado por `Advance(now)` e devolvendo eventos (RC4, RC5, RC7, RC8, RC11).
- [x] **T5** `Pick` e `Leave`: uma escolha por rodada, só na fase Choose e só de quem joga;
  revelar cedo quando todos escolheram ou quando quem faltava sai; Idle quando todos saem;
  placar ordenado com empate (RC9, RC11–RC15, RC18).
  - **Adicionado na T13:** "Jogar de novo" no placar inscreve o jogador na próxima partida. Sem
    isso, o E de quem ainda está no placar era ignorado.
- [x] **T6** `Room.SetBusy`: quem joga não anda, e o input é zerado ao sair (RC21).

## Fase 2 — Casos de uso e protocolo (Go)

- [x] **T7** `JoinCups`, `PickCup` e `LeaveCups`; o `RoomRunner` chama `Advance` em todo tick e
  traduz os eventos; `LeaveRoom` e a expiração tiram o jogador da partida. Steps godog com
  `FakeClock` e `Shuffler` fixo: **os 22 cenários passam** no godog.
- [x] **T8** Protocolo v1: mensagens `cups.*` no `adapter/ws`, `cups` no `room.joined` e testes
  que fixam o JSON exato de cada mensagem. Smoke `@smoke` com o servidor real.
- [x] **T9** `RW_CUPS_FAST=1` na config, dividindo todos os tempos por 10 (padrão: desligado).
  - **Mudou na T14:** o modo rápido mantém a escolha (10 s) e a revelação (2,5 s) no tempo real.
    Divididas por 10, seriam 1 s para clicar e 250 ms para ler "Acertou!", curto demais até
    para o teste.

## Fase 3 — Frontend (Vitest antes, navegador depois)

- [x] **T10** Protocolo TS das mensagens `cups.*` (mesmos JSONs dos testes Go), `nearBooth` e o
  reducer `cupsState`.
- [x] **T11** `timeline.ts`: fase e posição de cada copo pelo tempo decorrido, com o arco nas
  trocas; mesma resposta para qualquer `t`, inclusive depois de pular tempo (aba em segundo plano).
- [x] **T12** Na Praça (PixiJS): barraca e copos com `Graphics` e profundidade, copos seguindo a
  `timeline`, aviso "Aperte E para jogar", tecla E, copinho ao lado do nome e
  `data-cups` na lista acessível (RC2, RC20, RC22).
- [x] **T13** `CupsPanel.tsx`: espera com inscritos e contagem, copos como botões, teclas 1/2/3 e
  Esc (fora do chat), resultado da rodada, placar, **Jogar de novo** e **Sair** (RC6, RC8, RC12,
  RC13, RC19).
  - O aviso "Aperte E para jogar" fica embaixo, acima do chat, para não cobrir a barraca.
  - **Pedidos do dono (2026-10-26):** o painel da partida fica **no centro da tela**, e a
    espera (e o aviso de partida em andamento) fica **embaixo**, acima do chat. O sinal de quem
    joga é um copinho **antes** do nome (depois do nome, ficava colado nos copos da barraca).
  - Verificado no navegador com dois guaxinins e um robô: aviso, espera, rodada, escolha com a
    tecla 2, "Errou!" com os copos levantados, copos mexendo na barraca para quem assiste e o
    sinal ao lado do nome de quem joga.
  - Para os testes de navegador lerem o que a barraca faz (ela é desenhada no canvas), uma linha
    escondida `Barraca dos três copos` com `data-stage`, como a lista de jogadores.
- [x] **T14** Steps do playwright-bdd para os cenários `@e2e`, lendo `cups.round` pelo
  WebSocket para achar o copo certo, com o servidor em `RW_CUPS_FAST=1`.
  - **19/19 cenários de navegador passam** (14 das RW-001/002 e 5 da RW-005).
  - O andar até a barraca é em toques proporcionais à distância: segurando a tecla até chegar,
    o atraso de dois navegadores fazia o guaxinim passar ~100 px do ponto.
  - A troca de copos no modo rápido dura menos de 0,3 s: os steps esperam por ela checando a cada
    quadro (`waitForFunction` com `raf`).
  - Tempo máximo por cenário de 30 s para 90 s: dois jogadores abrindo o jogo levam ~40 s.

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
