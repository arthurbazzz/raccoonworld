# RW-006 — Plano técnico: Rummy das Nozes

Status: **Aprovado** (2026-10-30) · Spec: [spec.md](spec.md) · Criado: 2026-10-30

## 1. Visão geral

```
Navegador                                    Servidor Go
┌────────────────────────────────┐   WS   ┌────────────────────────────────────────┐
│ aviso "Aperte E" (perto)       │ ─────► │ adapter/ws: rummy.* (tiles_protocol)   │
│ lista de salas (React/DOM)     │        │        │                               │
│ painel da partida (React/DOM)  │ ◄───── │ usecase: OpenTable, CreateTableRoom,   │
│   mesa, mão, relógio, botões   │ eventos│   JoinTableRoom, MoveTiles, EndTurn,   │
│ mesa na Praça (PixiJS, sprite) │        │   DrawTile, UndoTurn, LeaveTableRoom   │
└────────────────────────────────┘        │        │                               │
                                          │ domain/rummy: peças, combinações,      │
                                          │   Table (salas) e Match (partida)      │
                                          │ RoomRunner: avança relógios no tick    │
                                          └────────────────────────────────────────┘
```

- Como na RW-005, **tudo vive no goroutine do `RoomRunner`** da Praça e avança no tick de
  20 Hz. Nada de locks nem timers próprios.
- Um pacote de domínio novo, `domain/rummy`, sem imports externos além de `domain/room`
  (para `PlayerID`).
- O navegador envia **intenções**; a mesa de rascunho do turno vive no servidor (seção 3.1).
- Nenhuma dependência nova: arrastar e soltar com a API de Pointer Events do DOM, e o `sharp`
  que já existe para recortar a arte.

## 2. Mesa no mapa (RN4, RN25)

- O `praca.json` ganha a mesa com tiles `M`, **sólidos**, nas colunas 13–16 das linhas 15–16
  (4×2 tiles, 128×64 px), abaixo da fonte e longe do spawn e dos obstáculos usados nos cenários
  da RW-001.
- A caixa da mesa sai dos tiles `M`, como a barraca sai dos `B`. `TileMap.NearTable` (Go) e
  `nearTable` (TS) reaproveitam a regra de "perto" da barraca (48 px), com uma tabela de casos
  comum.
- O desenho da mesa (128×96 px na arte) passa um pouco da caixa por cima, como as árvores.

## 3. Backend

### 3.1 Domínio: `domain/rummy`

**Peças** (`tile.go`)

| Tipo | Regra |
|---|---|
| `Tile` | `ID uint8` (0–105), `Color` (0–3: bolota, cereja, folha, gota), `Value` (1–13), `Joker bool`. O ID distingue as duas cópias |
| `NewDeck()` | As 106 peças em ordem fixa; o `Shuffler` embaralha |
| `Shuffler` | Interface `Shuffle([]Tile)`. Produção: `math/rand/v2`. Testes: ordem fixa |

**Combinações** (`meld.go`), funções puras e o núcleo dos testes unitários:

- `Classify(tiles []Tile) (MeldKind, error)`: `Group`, `Run` ou erro com motivo
  (`too_short`, `mixed_colors`, `repeated_color`, `gap`, `past_13`...).
- A **ordem** das peças importa só na sequência: o cliente manda a ordem em que estão na mesa.
  Coringas valem pela posição; um coringa num grupo vale o número do grupo.
- `Points(tiles)`: soma dos valores, com cada coringa valendo o número que substitui (RN13).

**Partida** (`match.go`)

| Conceito | Regra |
|---|---|
| `Match` | Modo, assentos (ordem dos turnos), mãos, monte, mesa **confirmada**, rascunho do turno, prazo do turno, passes seguidos |
| Rascunho | No começo da vez, o servidor copia mesa e mão. Cada `Move` altera **só o rascunho**. `EndTurn` valida o rascunho contra a cópia (RN11, RN13, RN14); se passar, vira a mesa confirmada |
| `Move(seat, layout)` | O cliente manda o **layout inteiro** do rascunho: lista de combinações (cada uma uma lista de IDs) e a mão. O servidor confere que é uma permutação das mesmas peças (nenhuma peça inventada ou sumida) e que nenhuma peça da mesa confirmada está na mão |
| `EndTurn(seat, now)` | Válido → confirma, passa a vez. Inválido ou nenhuma peça baixada → desfaz e compra (RN14). Mão vazia → fim (RN17) |
| `Draw(seat, now)` | Desfaz o rascunho e compra. Monte vazio → passe (RN16) |
| `Undo(seat)` | Rascunho volta à cópia (RN12) |
| `Advance(now)` | Prazo de 40 s vencido → como `Draw` (RN15). Ausentes por 30 s → saem (RN21). Devolve eventos |
| Saída | Peças do jogador voltam ao monte, embaralhadas (RN20). Se era a vez dele, o rascunho é descartado e a vez passa. Sobrou um → ele ganha (RN22) |
| Abertura | No modo Clássico, quem não abriu: as combinações da mesa confirmada precisam continuar idênticas no rascunho, e as novas, só com peças da mão, somam ≥ 30 |
| Placar | Pontos na mão de cada um, coringa = 30 (RN18, RN19) |

**Mesa e salas** (`table.go`)

| Conceito | Regra |
|---|---|
| `Table` | Lista de salas, cada uma com ID curto (`s1`, `s2`...), modo, jogadores (até 4), fase (`Lobby`, `Playing`, `Results`) e a `Match` quando joga |
| `Create(id, mode, now)` | Nova sala com o jogador. Quem já está numa sala não cria outra |
| `Join(id, room, now)` | Só sala em `Lobby` com vaga (RN7). Com 2 → contagem de 15 s; com 4 → começa já (RN6); voltar a 1 → contagem para |
| `Leave(id)` | Sai da sala ou da partida. Sala vazia some (RN5) |
| Assentos guardados | O assento é guardado pelo **apelido** (seção 3.2) |

Constantes num só arquivo: espera 15 s, vez 40 s, ausência 30 s, placar 8 s, abertura 30 pontos,
coringa na mão 30 pontos.

### 3.2 Desconexão e volta (RN21)

Hoje, um jogador desconectado sai da Praça em até 10 s (`PresenceTimeout`, RW-001) e, se
voltar, entra como um **jogador novo**, com outro `PlayerID`. Por isso:

- Quando um jogador sai da Praça no meio de uma partida, a `Match` marca o assento como
  **ausente**, guardando o **apelido** e o horário.
- Se alguém entra na Praça com o **mesmo apelido** em até 30 s desde a queda, o assento passa
  para o novo `PlayerID`, e ele recebe a mão e a mesa no `room.joined`. A RW-001 já impede dois
  jogadores com o mesmo apelido ao mesmo tempo.
- Depois de 30 s, o assento é liberado como em RN20.
- O guaxinim da volta nasce no spawn, e o painel abre sozinho. Não é preciso andar até a mesa.

Isso não muda a regra de 10 s da RW-001: só a mesa guarda o lugar.

### 3.3 Sala: travar o guaxinim (RN27)

Reaproveita o `busy` do `RoomRunner` (RW-005). Ocupado ao entrar numa sala da mesa, livre ao sair.
Quem está numa sala não pode entrar na barraca dos copos, e vice-versa.

### 3.4 Casos de uso

`OpenTable`, `CreateTableRoom`, `JoinTableRoom`, `MoveTiles`, `EndTurn`, `DrawTile`,
`UndoTurn` e `LeaveTableRoom`, no estilo do `JoinCups`.
`OpenTable` e `CreateTableRoom` conferem `NearTable` com a posição do servidor.
O `RoomRunner.step()` chama `table.Advance(now)` e traduz os eventos.
`JoinRoom` e `LeaveRoom` avisam a mesa (seção 3.2).

### 3.5 Protocolo WebSocket (v1, só adições)

**Cliente → servidor**

| type | data |
|---|---|
| `rummy.open` | `{}`: pede a lista de salas (tecla E perto da mesa) |
| `rummy.create` | `{ "mode": "classic"\|"free" }` |
| `rummy.join` | `{ "room": "s1" }` |
| `rummy.move` | `{ "turn": 7, "melds": [[3,17,40]], "hand": [5,9] }`: layout do rascunho. `turn` descarta movimentos atrasados |
| `rummy.end` / `rummy.draw` / `rummy.undo` | `{ "turn": 7 }` |
| `rummy.leave` | `{}` |

**Servidor → cliente**

| type | para | data |
|---|---|---|
| `rummy.rooms` | quem abriu a lista e quem está perto | `{ "rooms": [{ "id": "s1", "mode": "free", "players": [id], "phase": "lobby"\|"playing"\|"results", "lobbyMs": int }] }` |
| `rummy.status` | Praça | `{ "playing": [id], "tables": int }`: etiqueta "jogando" e quantas salas jogam (RN25, RN27) |
| `rummy.rejected` | jogador | `{ "reason": "too_far"\|"full"\|"playing"\|"busy" }` |
| `rummy.state` | **cada jogador da sala, separado** | `{ "room": "s1", "turn": 7, "current": id, "turnMs": int, "melds": [[...]], "hand": [ids], "counts": { id: 14 }, "pool": 78, "opened": [id], "absent": [id] }` |
| `rummy.draft` | os outros da sala | `{ "turn": 7, "melds": [[...]], "handCount": 12 }`: o rascunho de quem joga, ao vivo |
| `rummy.turn` | sala | `{ "turn": 8, "result": "ok"\|"invalid"\|"drew"\|"timeout"\|"passed", "reason": "...", "by": id }` |
| `rummy.results` | sala | `{ "winner": id\|null, "scores": [{ "id": id, "points": 23, "place": 2 }] }` |

- A peça viaja só pelo **ID** (0–105). O cliente sabe a cor e o número de cada ID pela mesma
  tabela fixa do `NewDeck`, testada nos dois lados com os mesmos dados.
- `rummy.state` é a única mensagem com `hand`, e cada jogador recebe só a sua (RN9, RN24). A
  ordem do monte nunca sai do servidor.
- **Rascunho ao vivo:** os outros veem as mexidas na hora (`rummy.draft`), o que é melhor de
  assistir e custa pouco: no máximo uma mensagem por soltar de peça. O rascunho mostra só as
  combinações e o tamanho da mão.

## 4. Frontend

### 4.1 Lógica pura (`src/game/rummy/`, Vitest)

- `tiles.ts`: tabela ID → cor e número.
- `melds.ts`: a mesma validação do Go, só para **destacar** combinações inválidas (RN12).
  Quem decide é o servidor. Os dois testes usam a mesma tabela de casos.
- `rummyState.ts`: reducer das mensagens `rummy.*`.
- `table.ts`: `nearTable`.

### 4.2 Painéis (React/DOM)

- `TableRoomsPanel.tsx`: lista de salas (modo, jogadores, vagas, "jogando") e "Criar sala" com
  a escolha do modo.
- `RummyPanel.tsx`: ocupa a maior parte da tela.
  - **Mesa**: combinações como linhas de peças; soltar uma peça entre duas combinações cria uma
    nova. Inválidas ficam com borda vermelha.
  - **Mão** embaixo, com "Ordenar por cor" e "Ordenar por número".
  - Relógio da vez, jogadores com o número de peças, e quem está ausente.
  - Botões **Terminar vez**, **Comprar**, **Desfazer** e **Sair**.
- Arrastar e soltar com Pointer Events (funciona com mouse; toque está fora de escopo). Cada
  peça é um elemento com `data-tile="17"` para os testes.

### 4.3 Arte (Q6)

- `npm run sprites` ganha um segundo passo que recorta `art/rummy.webp` com o `findFrames`
  que já existe:
  - Linha 1: verso marrom, monte marrom, 13 peças bolota, coringa.
  - Linhas 2–4: só as 13 peças coloridas. O verso e o monte coloridos são descartados.
  - O canto de baixo: a mesa e o verso marrom grande, ignorado.
- Saída em `public/sprites/rummy.png` com um manifesto (`rummy.json`): retângulo de cada peça
  por cor e número, coringa, verso, monte e mesa. Os dois coringas usam o mesmo desenho.
- Se a contagem de peças por linha não bater, o script para, como o do guaxinim.
- As peças do painel são `<div>` com a folha como `background-image`. A mesa da Praça é um
  sprite PixiJS na `scene.ts`.

## 5. Testes

| Camada | Ferramenta | Como |
|---|---|---|
| Cenários da spec | **godog** | Steps chamam os casos de uso com `FakeClock`, `FakeBroadcaster` e `Shuffler` fixo, para montar mãos conhecidas ("rocky tem o 7 azul") |
| Smoke | godog + servidor real | Criar sala, entrar, uma vez com compra, por WebSocket de verdade |
| Navegador | **playwright-bdd** | `@e2e`: aviso perto da mesa, criar e entrar em sala, arrastar peças e terminar a vez, destaque de inválida, Desfazer, mão escondida, duas salas ao mesmo tempo |
| Domínio | `go test` | `Classify` (tabela grande, foco em coringas), abertura, permutação do layout, fim por monte vazio, saída no meio, assento guardado |
| Frontend | Vitest | `melds`, `tiles`, reducer, protocolo e recorte da arte |

Para os testes de navegador, o servidor aceita `RW_RUMMY_FAST=1`, que reduz a espera para 1,5 s.
A vez continua com 40 s, para dar tempo de arrastar. Os testes de navegador também precisam de
mãos conhecidas: `RW_RUMMY_SEED=<n>` fixa o embaralhamento (só em teste).

## 6. Riscos

- **Validação de sequência com coringas** é o ponto mais sujeito a erro. Mitigação: uma tabela
  de casos compartilhada entre Go e TS.
- **Volta pelo apelido** confia que o apelido identifica o jogador. Sem contas (fora de escopo),
  alguém poderia pegar o apelido de quem caiu e ocupar o lugar dele. Aceitável enquanto não há
  prêmios.
- **Painel grande** cobre a Praça; o chat precisa continuar visível ao lado.
- **Arte gerada** pode mudar de tamanho numa nova geração; o recorte por detecção de bordas e a
  checagem de contagem pegam isso.

## 7. Decisões deste plano

1. **Pacote novo `domain/rummy`** no `RoomRunner` da Praça, como os copos.
2. **Rascunho no servidor**, com o cliente mandando o layout inteiro a cada mexida.
3. **Assento guardado pelo apelido** por 30 s, sem mudar a regra de 10 s da RW-001.
4. **Rascunho ao vivo** para os outros da sala.
5. **Mesa com tiles `M`** nas colunas 13–16, linhas 15–16.
6. **Recorte da arte** no `npm run sprites`, só com o verso marrom.
7. **`RW_RUMMY_FAST` e `RW_RUMMY_SEED`** só para testes.
