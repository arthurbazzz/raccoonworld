# RW-004 — Tasks: Guaxinim em sprites

Status: **Aprovado** (2026-10-25) · Plano: [plan.md](plan.md) · Criado: 2026-10-25

Regras: uma task por vez, teste antes do código, commits `RW-004/Tn: <descrição>`. A cada task,
os 14 cenários de navegador continuam verdes (RS1).

## Fase 0 — Arte e preparo

- [x] **T1** Versionar a imagem fonte em `web/art/raccoon.png`, declarar o `sharp` como
  devDependency e ignorar `web/public/sprites/` no git.
- [x] **T2** `scripts/sprites/slice.ts` (Vitest, TDD): achar fileiras e colunas pelas faixas
  transparentes, ignorar sujeira, gerar erro com a contagem quando não forem 3×6 (RS2).
- [x] **T3** Alinhamento pelos pés e montagem das células, ainda como lógica pura: tamanho da
  célula, `anchor`, `artHeight` e o conteúdo do JSON (RS3).
- [x] **T4** `scripts/sprites/build.ts` com `sharp` e o comando `npm run sprites`, ligado ao
  `predev`, `prebuild` e `pretest:e2e` (RS8). Rodar contra a imagem real e conferir as 18 células
  na saída.

## Fase 1 — Desenho no jogo

- [ ] **T5** `frameFor(facing, moving, t)`: fileira por direção, andar a 12 quadros/s e parado
  trocando a cada 1 s (RS5).
- [ ] **T6** `sheet.ts`: carregar o JSON e a imagem e cortar as texturas, com mipmaps.
  `RaccoonView` com `Sprite` de 46 px, espelhando só o sprite (RS1, RS4). Testes com texturas falsas.
- [ ] **T7** Cena e galeria usando a folha, com a mensagem de erro se ela não carregar. Remover
  a arte em código, a animação por código e os testes antigos (RS9).

## Fase 2 — Fechamento

- [ ] **T8** Rodar todas as suítes, conferir que o build falha com uma imagem fora do formato, pedir
  a **aprovação visual** do dono na galeria, conferir a Definição de Pronto e atualizar spec e plano.
