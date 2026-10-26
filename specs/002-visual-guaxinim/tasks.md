# RW-002 — Tasks: Visual do guaxinim

Status: **Aprovado** (2026-10-25) · Plano: [plan.md](plan.md) · Criado: 2026-10-25

Regras: as mesmas da RW-001. Uma task por vez, na ordem, teste antes do código, e commits
no formato `RW-002/Tn: <descrição>`.

## Fase 0 — Cenários

- [x] **T1** Escrever `specs/002-visual-guaxinim/features/visual.feature` com os 6 cenários da
  spec e a tag `@ui`, e criar os steps no playwright-bdd como pendentes.
  Pronto quando: `npm run test:e2e` lista os cenários novos **falhando**, e os 4 da RW-001
  continuam verdes.
  - Resultado: os 10 cenários da RW-002 (o esquema conta 4) falham com `pending: RW-002/T8`,
    e os 4 da RW-001 passam.
  - **Observação:** na primeira rodada, com o Next dev ainda frio, 2 cenários da RW-001 bateram
    no tempo limite de 1 min. Isolados, e numa segunda rodada completa, passaram. Isso deve ser
    avaliado na T9: rodar o e2e contra `next build && next start` em vez do modo dev.

## Fase 1 — Direção e estado (Vitest, TDD)

- [x] **T2** `facingFrom(dx, dy, previous)`: 4 lados, diagonal vale o lado, parado mantém e
  deslocamentos menores que 0,01 px contam como parado (RV1).
- [ ] **T3** `Interpolator.motion(id, now)`: `moving` nos últimos 150 ms, `facing` do último
  trecho com movimento, e quem acabou de entrar começa com `down` (RV1, RV2).

## Fase 2 — Arte em código

- [ ] **T4** `RaccoonView` com a paleta `RACCOON_COLORS` e as 3 poses (frente, costas e lado),
  com a esquerda espelhando só o corpo. Testes de montagem das partes por direção.
- [ ] **T5** Animações a partir do tempo: andar (bob, patas e rabo) e parado (respiração) (RV2).
- [ ] **T6** Anel no chão só no próprio guaxinim (RV3), sem mudar o ponto de apoio nem a hitbox (RV4).

## Fase 3 — Integração

- [ ] **T7** `RoomScene` passa a usar o `RaccoonView` e o `motion()`. O hook
  `useMotionAttributes` atualiza `data-facing`, `data-moving` e `data-ring` na lista acessível
  10 vezes por segundo.
- [ ] **T8** Implementar os steps de `visual.feature`. Pronto quando os 6 cenários passam.

## Fase 4 — Fechamento

- [ ] **T9** Rodar todas as suítes (RW-001 e RW-002), conferir a Definição de Pronto da
  constituição e atualizar spec e plan com o que mudou.
