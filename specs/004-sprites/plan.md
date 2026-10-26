# RW-004 — Plano técnico: Guaxinim em sprites

Status: **Implementado** (2026-10-26) · Aprovado: 2026-10-25 · Spec: [spec.md](spec.md) · Criado: 2026-10-25

## 1. Visão geral

```
web/art/raccoon.png  ──►  npm run sprites  ──►  web/public/sprites/raccoon.png   (folha alinhada)
 (imagem da IA,            (Node + sharp,        web/public/sprites/raccoon.json  (quadros + âncora)
  versionada)               roda no dev/build)                │
                                                              ▼
                                      jogo: carrega a folha uma vez → RaccoonView (Sprite)
```

Só o frontend muda. Direção, "andando", anel, lista acessível e os 14 cenários de navegador
ficam como estão (RS1).

## 2. Preparo da arte (RS2, RS3, RS6, RS8)

**Arquivos:**
- `web/art/raccoon.png`: a imagem **fonte**, entregue pelo dono. Ela é versionada.
- `web/scripts/sprites/slice.ts`: lógica pura. Recebe os pixels RGBA e devolve os quadros. Testada com Vitest.
- `web/scripts/sprites/build.ts`: lê e grava as imagens com o `sharp` e chama o `slice`.

Os dois rodam com o **TypeScript nativo do Node 24** (`node scripts/sprites/build.ts`), sem compilar.

**Detecção** (`slice.ts`):
1. Um pixel é "cheio" quando o alfa passa de 128.
2. As **fileiras** são faixas horizontais com pixels cheios, separadas por linhas 100% transparentes.
   Faixas com menos de 8 px são descartadas como sujeira.
3. Dentro de cada fileira, as **colunas** são achadas do mesmo jeito, na vertical.
4. Para cada quadro, guarda a caixa justa (`x0, x1, top, bottom`), o centro dos pés
   `(x0+x1)/2, bottom` e a altura.
5. Se não forem exatamente **3 fileiras × 6 colunas**, gera um erro com o que achou. Por exemplo:
   `"esperado 3×6 quadros, achei 3 fileiras com 6, 5, 6 colunas"`.

**Saída** (`build.ts`):
- A **folha alinhada** `public/sprites/raccoon.png` tem 6×3 células do mesmo tamanho. A célula é a
  maior largura e a maior altura entre os quadros, mais 4 px de margem.
- Todo quadro é colado na célula com os **pés no mesmo ponto** (`anchor`): no centro e embaixo.
  Isso corrige o ±2 px da IA.
- O **`raccoon.json`** tem esta forma:

```json
{
  "image": "/sprites/raccoon.png",
  "cell": { "w": 112, "h": 120 },
  "anchor": { "x": 56, "y": 116 },
  "rows": { "front": 0, "back": 1, "side": 2 },
  "idle": [0, 1],
  "walk": [2, 3, 4, 5],
  "artHeight": 110
}
```

`artHeight` é a mediana das alturas, usada para a escala (RS4).

**Quando roda** (RS8):
- `npm run sprites` também é chamado pelo `predev`, `prebuild` e `pretest:e2e`.
- Os arquivos gerados em `public/sprites/` **não são versionados** (`.gitignore`), para não haver
  cópia velha da arte.
- Se a imagem estiver errada, `dev`, `build` e `test:e2e` **param** com a mensagem do passo 5.

## 3. Desenho no jogo

**`web/src/game/raccoon/sheet.ts`:**
- `loadRaccoonSheet()` lê o JSON e a imagem pelo `Assets` do PixiJS e corta uma `Texture` por
  quadro (`Texture` com `frame`), em `frames[row][col]`. É chamado **uma vez** pela cena e pela galeria.
- Os quadros são reduzidos para cerca de 0,42×. Por isso a textura usa **mipmaps** e filtro linear,
  para não serrilhar.

**`frameFor(facing, moving, t)`**, uma função pura:
- fileira: `down`→front, `up`→back, `left`/`right`→side;
- andando: `walk[floor(t · 12 / 1000) mod 4]` (Q3, `WALK_FPS = 12`);
- parado: `idle[floor(t / 1000) mod 2]`, ou seja, troca de quadro a cada 1 s, como uma respiração.

**`RaccoonView`** mantém a interface que a cena usa (`setFacing`, `animate`, `ring`, `facing`):
- recebe a folha no construtor: `new RaccoonView(sheet, { isYou })`;
- o corpo é um `Sprite` com `anchor` = `anchor / cell` e escala `TARGET_HEIGHT / artHeight`,
  com `TARGET_HEIGHT = 46` (RS4);
- **olhar para a esquerda:** `scale.x` negativo **só no sprite**. O anel, a sombra e o nome não espelham;
- `animate(moving, t)` só troca a textura por `frameFor`. O movimento do andar já está desenhado
  nos quadros, então o balanço, as patas e o rabo feitos em código saem.

**Remoção** (RS9):
- saem `palette.ts`, `shapes.ts`, `rig.ts`, a arte antiga dentro do `RaccoonView.ts` e os testes
  dessas peças (`legacyArt`, `rig`, `shapes`, `palette`);
- também saem `animationFrame`, `STEPS_PER_SECOND` e `BREATH_PERIOD_MS`, substituídos por `frameFor`.

**Cena e galeria:**
- `RoomScene.create` carrega a folha antes de criar os guaxinins. Se falhar, aparece a mensagem
  de erro que o `GameCanvas` já mostra;
- a galeria usa a mesma folha (RS7).

## 4. Testes

| O quê | Ferramenta | Como |
|---|---|---|
| Detecção (RS2) | Vitest | Imagens de teste geradas no próprio teste (buffers RGBA): grade certa, fileiras fora da grade de 128, sujeira pequena ignorada |
| Erro de formato (RS2) | Vitest | Imagem com 5 colunas numa fileira → erro com a contagem |
| Alinhamento (RS3) | Vitest | Quadros com pés em alturas diferentes saem com os pés no mesmo `anchor` |
| `frameFor` (RS5) | Vitest | Fileira por direção, os 4 quadros de andar a 12/s e os 2 de parado a cada 1 s |
| `RaccoonView` (RS1, RS4) | Vitest | Com texturas falsas: a textura por direção e tempo, o espelho só no sprite, o anel, a altura de ~46 px |
| Arte real | `npm run sprites` | Roda contra `web/art/raccoon.png` e precisa achar 3×6 |
| Comportamento (RS1) | playwright-bdd | Os 14 cenários, sem mudança |
| Visual | Dono do projeto | Aprovação na `/galeria` |

## 5. Dependências novas

| Dependência | Motivo |
|---|---|
| `sharp` (devDependency) | Ler e gravar PNG no script de preparo. Hoje ele já vem instalado indiretamente pelo Next, mas o projeto passa a usá-lo diretamente, então ele fica declarado. |

## 6. Riscos

- **Arte nova fora do formato:** o build falha de propósito, com a mensagem. A saída é corrigir a
  imagem (ou pedir de novo à IA).
- **Tamanho:** a folha alinhada tem ~670×360 px (~100 KB). Ela é carregada depois do login, junto
  com o PixiJS, então a primeira tela não muda.
- **TypeScript nativo do Node:** os scripts não podem usar sintaxe que precise de compilação
  (`enum` e parâmetros de construtor com modificador). Isso é registrado no CLAUDE.md.

## 7. O que mudou durante a implementação

| Mudança | Task | Por quê |
|---|---|---|
| A montagem da grade ficou em `scripts/sprites/layout.ts`, separada do `build.ts` | T3 | Lógica pura, testável sem imagens |
| O tipo `SheetManifest` fica em `src/game/raccoon/manifest.ts` e é importado pelos scripts | T3 | Uma só definição para quem escreve e quem lê o JSON |
| `web/scripts/package.json` com `"type": "module"` | T4 | O Node avisava que não sabia se os scripts eram ESM |
| `tsconfig`: `allowImportingTsExtensions` | T2 | O Node exige imports com `.ts` para rodar TypeScript nativo |
| A escolha do quadro ficou em `frames.ts` (`frameFor`, `WALK_FPS`, `IDLE_FRAME_MS`) | T5 | Separada do desenho, testável sem PixiJS |
| O novo `RaccoonView` nasceu em `RaccoonSprite.ts`; o antigo `RaccoonView.ts` foi apagado na T7 | T6/T7 | O jogo não quebra entre as tasks |
| Galeria com zoom 2,5× (era 4×) | T7 | Acima da resolução da arte o sprite borra |
| Steps do e2e esperam o canvas aparecer depois de entrar | T7 | Com o Next frio, "Animação de andar" disputava com a página ainda carregando |
| A folha gerada tem ~285 KB (o plano estimava ~100 KB) | T4 | PNG com alfa e sombreado suave; ainda carrega só depois do login |

## 8. Decisões deste plano (aprovadas em 2026-10-25)

1. A imagem **fonte** fica em `web/art/` e é versionada, e os arquivos **gerados** não são versionados.
2. O `sharp` passa a ser uma `devDependency` declarada.
3. Os scripts rodam com o **TypeScript nativo do Node**, sem uma ferramenta extra.
4. O **balanço, as patas e o rabo animados por código saem**, porque o movimento já vem nos quadros.
5. O guaxinim fica com **46 px de altura** no jogo.
