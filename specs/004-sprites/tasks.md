# RW-004 — Tasks: Guaxinim em sprites

Status: **Concluído** (2026-10-26), menos a aprovação visual · Aprovado: 2026-10-25 · Plano: [plan.md](plan.md) · Criado: 2026-10-25

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

- [x] **T5** `frameFor(facing, moving, t)`: fileira por direção, andar a 12 quadros/s e parado
  trocando a cada 1 s (RS5).
- [x] **T6** `sheet.ts`: carregar o JSON e a imagem e cortar as texturas, com mipmaps.
  `RaccoonView` com `Sprite` de 46 px, espelhando só o sprite (RS1, RS4). Testes com texturas falsas.
- [x] **T7** Cena e galeria usando a folha, com a mensagem de erro se ela não carregar. Remover
  a arte em código, a animação por código e os testes antigos (RS9).
  - A cena carrega a folha em `RoomScene.create`. Saíram 9 arquivos da RW-002/RW-003:
    `RaccoonView.ts`, `palette`, `shapes`, `rig` e os testes deles.
  - Galeria com zoom 2,5× (perto da resolução da arte); com 4× o sprite ficava borrado.
  - `public/sprites/` entrou no `.prettierignore`, porque é gerado.
  - **E2E:** "Animação de andar" falhou 1 vez em 5 com o Next frio. O rocky bate no muro de baixo
    em ~1,3 s, e a página do observador ainda carregava o PixiJS e o sprite. O `join` dos steps
    agora espera o canvas aparecer. Depois disso, 3/3 a frio e a suíte completa 14/14. Cada
    cenário com 2 jogadores passou de ~4 s para ~17–20 s, porque agora espera o jogo carregar
    em contextos sem cache; no navegador, o canvas aparece ~0,4 s depois de "Entrar".

## Fase 2 — Fechamento

- [x] **T8** Rodar todas as suítes, conferir que o build falha com uma imagem fora do formato, pedir
  a **aprovação visual** do dono na galeria, conferir a Definição de Pronto e atualizar spec e plano.

## Fechamento (T8, 2026-10-26)

### Definição de Pronto

- [x] **Critérios de aceitação:**
  - **14/14 cenários de navegador** (RW-001 e RW-002) com o guaxinim em sprites, rodados a frio;
  - testes de detecção, erro, alinhamento, escolha de quadro e do `RaccoonView`.
- [ ] **Aprovação visual do dono na `/galeria`:** pendente.
- [x] **Testes, lint e formatação:**
  - Vitest com 134 testes, `tsc`, ESLint, Prettier e `next build`;
  - Go: `go vet`, golangci-lint com 0 problemas e todos os pacotes verdes, sem mudanças nesta RW.
- [x] **Camadas:** o servidor não mudou, e o preparo da arte fica em `web/scripts/`, fora do jogo.
- [x] **spec, plan e tasks atualizados:** o plan §7 lista o que mudou.

### Conferências da RW-004

- **Imagem fora do formato (RS2/RS8):** com a última coluna da fileira de lado apagada,
  `npm run build` parou com "esperado 3×6 quadros, achei 3 fileiras com 6, 6, 5 colunas", e o
  `next build` não rodou. A imagem original foi restaurada.
- **Produção:** `/` responde 200, `/galeria` responde 404, `/sprites/raccoon.json` responde 200 (283 B)
  e `/sprites/raccoon.png` responde 200 (285 KB). A primeira carga continua em 103–104 kB de JS.
- **No jogo:** o rocky aparece como sprite com o anel, vira para a esquerda ao andar e para
  olhando para a esquerda. O canvas aparece ~0,4 s depois de "Entrar".

### Observações

- **O andar de frente é sutil na arte:** as pernas pouco mudam entre os quadros. Trocar a imagem
  resolve sem mudar código (RS6).
- **Os cenários de navegador ficaram mais lentos** (~17–20 s com 2 jogadores), porque agora esperam
  o jogo carregar em contextos sem cache, no modo dev.
