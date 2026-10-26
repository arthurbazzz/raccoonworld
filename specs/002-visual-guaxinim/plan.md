# RW-002 — Plano técnico: Visual do guaxinim

Status: **Aprovado** (2026-10-25) · Spec: [spec.md](spec.md) · Criado: 2026-10-25

## 1. Visão geral

Só o frontend muda. Servidor, protocolo e colisão ficam como estão (spec, "Fora de escopo").

```
snapshots ──► Interpolator ──► motion(id, now) ──► { facing, moving }
                    │                                   │
                    └── position(id, now) ──► RaccoonView.update(pos, facing, moving, now)
                                                        │
                                         lista acessível (data-facing, data-moving, data-ring)
```

- A **direção** e o **andando/parado** saem das mesmas posições interpoladas que a cena já
  desenha. Assim o guaxinim vira exatamente quando começa a andar na tela, e todo mundo vê
  o mesmo (RV5), porque as posições vêm do servidor.
- O desenho fica num componente próprio (`RaccoonView`), separado da cena.

## 2. Movimento → direção e estado (`web/src/game/motion.ts`)

Funções puras, testadas com Vitest:

```ts
type Facing = "down" | "up" | "left" | "right";
interface Motion { facing: Facing; moving: boolean }

function facingFrom(dx: number, dy: number, previous: Facing): Facing
```

- `dx ≠ 0` → `left`/`right`; senão `dy > 0` → `down`, `dy < 0` → `up`; parado → mantém
  `previous`. Na diagonal, `dx ≠ 0`, então vale o lado (RV1).
- Deslocamentos menores que 0,01 px contam como parado, para ruído de ponto flutuante não virar o guaxinim.
- O `Interpolator` ganha `motion(id, now)`:
  - `moving` = a posição interpolada mudou nos últimos **150 ms** (RV2);
  - `facing` = a direção do último trecho em que houve movimento;
  - quem acabou de entrar começa com `down` (RV1).

## 3. Arte em código (`web/src/game/raccoonArt.ts`)

`RaccoonView` é um `Container` do PixiJS com as partes separadas, para poder animar cada uma:

| Parte | Frente (`down`) | Costas (`up`) | Lado (`left`/`right`) |
|---|---|---|---|
| Corpo | oval cinza com barriga clara | oval cinza, sem barriga | oval mais comprido |
| Cabeça | rosto com máscara, olhos e focinho | nuca e orelhas | perfil, focinho para a frente |
| Rabo | aparece atrás, de lado | em destaque, listrado | atrás do corpo, listrado |
| Patas | 4 patinhas | 4 patinhas | 2 de cada lado, alternando |

- **Esquerda é a direita espelhada** (`scale.x = -1` no corpo). O nome e o anel não são espelhados.
- As **cores** viram uma paleta única (`RACCOON_COLORS`), já pensando na customização futura.
- O ponto de apoio continua nos pés `(0, 0)`, e a ordenação 3/4 e a hitbox não mudam (RV4).

**Animações**, calculadas a cada quadro a partir do tempo (sem sprite sheet):

| Estado | Efeito |
|---|---|
| Andando | corpo sobe e desce (bob de 2 px, ~6 passos/s), patas alternadas, rabo balança |
| Parado | respiração: corpo escala 1,00 → 1,03 em Y, ciclo de ~2 s |

- **Anel (RV3):** elipse com contorno amarelo nos pés, desenhada antes do corpo, só no
  guaxinim do próprio jogador.

## 4. Cena e lista acessível

- `RoomScene.place()` passa a chamar `view.update(pos, motion, now)` em vez de só mover o `Container`.
- A lista oculta de jogadores ganha `data-facing`, `data-moving` e `data-ring`. Esses valores
  vêm do mesmo `interpolator.motion()` que a cena usa, atualizado **10 vezes por segundo**
  por um hook (`useMotionAttributes`), para não renderizar o React a cada quadro.

## 5. Testes

| Camada | Ferramenta | O quê |
|---|---|---|
| Direção e estado | Vitest | `facingFrom` (4 lados, diagonal, parado), `motion()` (150 ms, primeira direção) |
| Arte | Vitest | `RaccoonView` monta as partes certas para cada direção, espelha só o corpo, anel só no próprio |
| Cenários da spec | playwright-bdd | `specs/002-visual-guaxinim/features/visual.feature` com a tag `@ui`, lida pela mesma config do Playwright |
| Regressão | Vitest + playwright-bdd | As suítes da RW-001 continuam verdes |

O godog não lê esta pasta: não há nada novo no servidor.

## 6. Dependências novas

Nenhuma.

## 7. Riscos

- **Desempenho:** redesenhar Graphics a cada quadro com 10 guaxinins custa pouco. Mesmo
  assim, as poses são montadas uma vez por mudança de direção, e por quadro só mudam
  posição, escala e rotação.
- **Gosto:** arte em código tem um teto de qualidade. A paleta e as partes separadas
  facilitam trocar por sprites depois, sem mexer em direção e animação.

## 8. Decisões deste plano (aprovadas em 2026-10-25)

1. **Direção e estado deduzidos das posições interpoladas** (o que aparece na tela), e não
   das teclas. Isso vale também para o próprio jogador, que continua seguindo o servidor
   sem predição.
2. Lista acessível atualizada **10 vezes por segundo**, para os testes e para leitores de tela.
3. Os cenários da RW-002 rodam **só no navegador** (`@ui`).
