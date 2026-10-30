# RW-006 — Prompts para gerar a arte

Estilo de referência: o guaxinim de `web/art/raccoon.png` (pixel art, contorno escuro, tons
quentes). Gere cada imagem com **fundo transparente** e as peças **alinhadas numa grade**, para
o script de sprites recortar (como na RW-004). O formato exato da grade é definido no plan.

## 1. Folha de peças

```
Pixel art sprite sheet of game tiles for a cozy raccoon-themed tile rummy game.
Transparent background. A strict grid of 14 columns x 4 rows, each cell 48x64 px,
no gaps, no text outside the tiles.
Each tile: a cream-colored rounded rectangle with a dark brown 2px outline and a
slight bottom shadow, like a small wooden board.
Row 1 = acorn brown, row 2 = berry red, row 3 = moss green, row 4 = river blue.
Columns 1-13: a large, bold, centered number from 1 to 13 in the row's color,
with a tiny matching icon below it (acorn, berry, leaf, water drop).
Column 14 of every row: the joker tile, a cream tile with a smiling raccoon mask
face instead of a number.
Consistent lighting, crisp pixels, no anti-aliasing, no blur.
```

## 2. Verso da peça e monte

```
Pixel art, transparent background, two sprites side by side, each 48x64 px:
(1) the back of a game tile: dark wooden board with a dark brown outline and a
small raccoon paw print in the center;
(2) a small stack of five face-down tiles seen from the front, slightly messy.
Same style as a cozy raccoon village game, crisp pixels, no anti-aliasing.
```

## 3. Mesa na Praça

```
Pixel art top-down 3/4 view of a round wooden picnic table for a cozy raccoon
village game, transparent background, 128x96 px. A few cream game tiles lie
scattered on top, with a small tree stump stool on each of the four sides.
Warm colors, dark brown outline, crisp pixels, no anti-aliasing.
```

## Observações

- As 4 cores precisam ser distinguíveis por quem tem daltonismo: o ícone de cada cor (bolota,
  fruta, folha, gota) garante isso.
- Se a ferramenta não respeitar a grade exata, gere uma peça por imagem e monte a folha depois.
