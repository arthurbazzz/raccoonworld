# RW-003 — Tasks: Arte do guaxinim em camadas

Status: **Aprovado** (2026-10-25) · Plano: [plan.md](plan.md) · Criado: 2026-10-25

Regras: as mesmas das RWs anteriores. Uma task por vez, teste antes do código, e commits
`RW-003/Tn: <descrição>`. A cada task, os 14 cenários de navegador e os testes unitários de
comportamento da RW-002 continuam verdes (RA1).

## Fase 0 — Estrutura

- [x] **T1** Criar `web/src/game/raccoon/`, mover o `RaccoonView` atual para lá sem mudar o
  desenho e deixar `raccoonArt.ts` só reexportando. Separar os testes da RW-002 em
  "comportamento", que continuam, e "desenho antigo", que serão trocados nas próximas tasks.
  Pronto quando: tudo verde, sem mudança visual.
  - O `raccoonArt.ts` **não ficou** como reexportação: o único importador era a `scene.ts`, que
    agora importa de `./raccoon`. O plano já previa a remoção quando ninguém mais importasse.
  - Os testes ficaram em `raccoon/RaccoonView.test.ts` (21, comportamento) e
    `raccoon/legacyArt.test.ts` (2, desenho antigo: quantidade de patas e chaves da paleta).
- [x] **T2** Galeria `/galeria` (RA8) com o guaxinim atual: 4 direções, andando e parado, a 1× e a 4×,
  respondendo 404 em produção. Ela serve de "antes" para comparar com o resultado.
  - Layout: uma fileira por estado (parado, andando), com as 4 direções a 4× e o tamanho real
    embaixo; o primeiro de cada fileira tem o anel. Cabe em ~900 px de largura.
  - Conferido com `next build && next start`: `/` responde 200 e `/galeria` responde 404.
  - **Cuidado:** rodar `next build` e depois `next dev` na mesma pasta deixou o dev servindo
    arquivos inexistentes (404 nos chunks). Apagar `web/.next/` resolve.

## Fase 1 — Paleta e formas (Vitest, TDD)

- [x] **T3** `palette.ts`: paleta base, `shade()` e contorno derivados (RA5, RA7).
- [x] **T4** `shapes.ts`: formas com contorno e sombra pela técnica das 2 formas (elipse, círculo,
  orelha e rabo com anéis), todas recebendo a paleta.

## Fase 2 — Poses novas

- [ ] **T5** Pose de **frente** no estilo da referência: cabeça grande, máscara, faixa creme, olhos
  com brilho, nariz, boquinha, braços, pernas, barriga e rabo. Camadas na ordem do plano (RA2).
- [ ] **T6** Pose de **lado**: perfil com focinho, perna de trás e da frente, braço e rabo atrás.
- [ ] **T7** Pose de **costas** com mais detalhe (Q2): nuca clara, orelhas com contorno e rabo com
  volume por cima do corpo.
- [ ] **T8** Animação sobre as partes novas: braços em oposição às pernas, e o resto como na RW-002.
  Conferir a legibilidade a 1× na galeria (RA6).

## Fase 3 — Encaixes e acessórios

- [ ] **T9** `SlotAnchor` por pose no `rig.ts`, mais `equip`/`unequip` no `RaccoonView`, com
  reequipar na troca de pose (RA3). Testes com um acessório criado só no teste (RA4).
- [ ] **T10** Chapéu de exemplo em `accessories.ts`, nas 3 poses, mostrado na galeria com e sem
  ele (Q4).

## Fase 4 — Fechamento

- [ ] **T11** Rodar todas as suítes, conferir o 404 da galeria com `next build && next start`,
  pedir a **aprovação visual** do dono na galeria, conferir a Definição de Pronto e atualizar
  spec e plan com o que mudou.
