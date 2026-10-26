# RW-002 — Tasks: Visual do guaxinim

Status: **Concluído** (2026-10-25) · Aprovado: 2026-10-25 · Plano: [plan.md](plan.md) · Criado: 2026-10-25

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
- [x] **T3** `Interpolator.motion(id, now)`: `moving` nos últimos 150 ms, `facing` do último
  trecho com movimento, e quem acabou de entrar começa com `down` (RV1, RV2).

## Fase 2 — Arte em código

- [x] **T4** `RaccoonView` com a paleta `RACCOON_COLORS` e as 3 poses (frente, costas e lado),
  com a esquerda espelhando só o corpo. Testes de montagem das partes por direção.
- [x] **T5** Animações a partir do tempo: andar (bob, patas e rabo) e parado (respiração) (RV2).
- [x] **T6** Anel no chão só no próprio guaxinim (RV3), sem mudar o ponto de apoio nem a hitbox (RV4).

## Fase 3 — Integração

- [x] **T7** `RoomScene` passa a usar o `RaccoonView` e o `motion()`. O hook
  `useMotionAttributes` atualiza `data-facing`, `data-moving` e `data-ring` na lista acessível
  10 vezes por segundo.
- [x] **T8** Implementar os steps de `visual.feature`. Pronto quando os 6 cenários passam.
  - Resultado: os 10 cenários da RW-002 passam (os 6 da spec, com o esquema contando 4), e os 4 da
    RW-001 também. A suíte completa passou 2 vezes seguidas.
  - Mutações: "nunca andando" foi pega. "Diagonal pela frente/costas" **escapou** duas vezes, por
    dois motivos, ambos corrigidos:
    1. com W+D, o guaxinim bate na fonte e passa a andar reto, então o cenário virou **S+D**
       (spec e `.feature` atualizados);
    2. a nova tentativa de 5 s do Playwright esperava o guaxinim chegar ao muro e andar reto, então
       a verificação de direção agora espera no máximo 1 s.
    Depois disso, a mutação passou a ser pega.
  - O anel é verificado pelo `data-ring`, que vem do mesmo `you` que a cena usa. O desenho do anel
    em si é coberto pelos testes do `RaccoonView` (T6).

## Fase 4 — Fechamento

- [x] **T9** Rodar todas as suítes (RW-001 e RW-002), conferir a Definição de Pronto da
  constituição e atualizar spec e plan com o que mudou.

## Fechamento (T9, 2026-10-25)

### Definição de Pronto

- [x] **Critérios de aceitação da spec atendidos.** Os 6 cenários de `features/visual.feature` passam.
- [x] **Cenários BDD passando:**
  - playwright-bdd: **14/14**, sendo 10 da RW-002 e 4 da RW-001;
  - godog: **24/24** nos casos de uso e **3/3** no smoke. A RW-002 não muda o servidor, então
    essas suítes só confirmam que nada quebrou.
- [x] **Testes, lint e formatação limpos:**
  - Vitest com 127 testes (+50 desde a RW-001), `tsc`, ESLint, Prettier e `next build`;
  - Go: `go vet`, golangci-lint com 0 problemas e `golangci-lint fmt` sem diferenças.
- [x] **Nenhuma violação de camada.** Nenhum arquivo fora de `web/` e `specs/` mudou, e o servidor
  está idêntico ao da RW-001, como a spec pedia.
- [x] **spec, plan e tasks atualizados.** O plan §8 lista o que mudou.

As mutações confirmaram que os cenários pegam os erros: "nunca andando" e "diagonal pela
frente/costas", esta depois das correções da T8.

### Observações

- A primeira carga da página continua em ~108 kB, porque o PixiJS e a arte carregam depois do login.
- **O e2e roda contra o Next em modo dev.** Na T1, 2 cenários estouraram o tempo limite com o
  servidor frio. Nas 5 rodadas completas seguintes isso não se repetiu, então mantive o modo dev.
  Se voltar a acontecer, a saída é rodar o e2e contra `next build && next start`.
- A fluidez da animação não foi avaliada a olho, porque o painel do navegador usado nos testes
  desenha a 2 quadros por segundo. Vale conferir com o jogo numa janela visível.

### Próximas specs sugeridas

- Customização do guaxinim (cores e acessórios), que troca a `RACCOON_COLORS` e depende da autenticação.
- Arte do cenário (chão, árvores e fonte) no mesmo estilo desenhado em código.
