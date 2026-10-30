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

## 8. O que mudou durante a implementação

| Mudança | Task | Por quê |
|---|---|---|
| Cores dos cenários: marrom, vermelho, verde e azul (em vez de preto e laranja) | T1 | São as cores da arte entregue (Q6) |
| Frases dos cenários diferentes das dos copos ("cria uma sala", "sai da mesa", "sai da Praça") | T1 | Os steps dos copos já usam "aperta a tecla E/Esc" e "a conexão cai" |
| `rummy.rooms` vai para toda a Praça a cada mudança, e não só para quem abriu a lista | T10 | Mais simples e pequeno; o cliente só mostra a lista com o painel aberto |
| Mensagem nova `rummy.again` (Jogar de novo) | T11 | O plano previa o botão, mas não a mensagem |
| `rummy.state` ganhou `mode`; `rummy.turn` ganhou `drew`; `cups.rejected` ganhou o motivo `busy` | T10, T11 | O painel precisa do modo e de saber se houve compra; os copos recusam quem está na mesa |
| O "ocupado" do runner virou dois conjuntos, copos e rummy | T10 | Uma mudança na barraca liberava quem estava jogando rummy |
| A arte sai como PNG com o manifesto apontando os recortes na imagem original, sem remontar a folha | T13 | As peças são retângulos simples; não há pés a alinhar como no guaxinim |
| Corte da tira de peças nas colunas escuras do contorno | T13 | O entalhe do coringa também é um vale de opacidade e virava uma 15ª peça |
| T5 a T8 num commit só | T5–T8 | As quatro ficaram no mesmo arquivo (`match.go`) |

## 9. Emenda A (RN12a, RN14–RN15, RN28–RN39)

Status: **Aprovado** (2026-10-30) · Spec: Emenda A aprovada em 2026-10-30.

### 9.1 Jogada recusada não encerra a vez (RN14, RN15)

- `Match.EndTurn` deixa de chamar `drawAndPass` quando a jogada é inválida. O rascunho continua
  como está, e o resultado é um novo evento `Refused{Player, Issues}`, enviado **só ao jogador
  da vez**. A vez não muda e o relógio continua.
- Quando o tempo acaba, `Advance` desfaz e compra, como hoje. Com "nada baixado", o servidor
  recusa sem comprar. O cliente já desabilita o botão nesse caso.
- `rummy.turn` com `result: "invalid"` deixa de existir.

### 9.2 Erros claros (RN12a)

- `Classify` passa a devolver um `Issue{Code, Tiles []int, Params}`, além do tipo de combinação:
  - `Tiles` são as peças culpadas (as duas da mesma cor, a de cor diferente, a vizinha do buraco);
  - `Params` traz o que falta para montar a frase: cor e número repetidos, número faltando,
    pontos que faltam na abertura e a peça da mesa usada na abertura.
- O servidor manda `Issues []{meld, code, tiles, params}`. O cliente monta a frase em
  português com `describeIssue`. O texto fica no cliente, e o servidor só manda códigos.
- O TypeScript (`melds.ts`) faz a mesma conta, para mostrar o erro ao vivo. Os casos novos
  entram em `testdata/melds.json`, que já é usado pelo Go e pelo TS, e isso garante que os dois
  concordam.

### 9.3 Posição das combinações (RN32–RN36)

- `Meld` vira `{tiles []int, x, y int}` em unidades de grade da mesa (48×28 células). O cliente
  manda o layout inteiro em `rummy.move`, como hoje, e agora com as posições.
- O servidor só confere os limites da mesa e repassa as posições no rascunho (`rummy.draft`).
  A regra de não sobrepor é resolvida no cliente: ele empurra a combinação para a vaga livre
  mais perto (`layout.ts`, função pura).
- **Dividir, juntar e várias peças** são operações puras em `layout.ts` (`insertAt`, `cutAt`,
  `joinEnds`, `moveMany`) que geram o próximo layout. O servidor não precisa saber qual foi o
  gesto.
- No modo Clássico antes de abrir, o servidor já recusa mexer nas combinações da mesa. Mudar
  só a posição continua permitido, porque não altera as peças.
- Desfazer volta às posições do começo da vez, que já ficam guardadas em `startMelds`.

### 9.4 Chat da mesa (RN37–RN39)

- Novas mensagens: `rummy.chat.send {text}` e `rummy.chat {from, text}`. Elas vão só para os
  jogadores da sala.
- O tamanho de 120 caracteres e o limite de envio usam o mesmo limitador do `chat.send`, por
  jogador.
- O histórico fica só no cliente e é apagado no fim da partida. O servidor não guarda nada.

### 9.5 Frontend: cena da mesa (RN28–RN31)

- `RummyPanel` dá lugar a `RummyTableScene`, uma camada em tela cheia feita em React/DOM (não
  em Pixi), por cima da Praça. É DOM porque arrastar, selecionar e mostrar balões são mais
  simples e testáveis assim. Ela tem:
  - `TableFelt`: mesa oval com as combinações em posição absoluta pela grade (9.3), o monte
    num canto e o retângulo de seleção;
  - `OpponentSeat`: avatar do guaxinim (o sprite da Praça), apelido, número de peças, anel da
    vez em SVG (`stroke-dashoffset` a partir de `secondsUntil`) e o balão de chat. Os lugares
    são em cima, à esquerda e à direita, conforme quantos jogadores há;
  - `TileRack`: suporte de 2 fileiras × 14 vagas, desenhado em CSS como um porta-peças escuro;
  - `TableChat`: painel recolhível à direita;
  - a barra de ações ao lado do suporte.
- **Suporte** (`rack.ts`, função pura, Vitest): um mapa local `vaga → peça`.
  - O suporte se ajusta à mão que vem do servidor: peças novas entram na primeira vaga livre e
    peças que saíram liberam a vaga.
  - "Por cor" e "Por número" refazem o mapa.
  - Nada disso vai para o servidor.
- **Seleção** (`selection.ts`): Shift+clique, retângulo e Esc. Ao arrastar, todas as peças
  marcadas vão juntas.
- A `TableRoomsPanel` (lista de salas e espera) continua como está.

### 9.6 Testes

- **godog:** os cenários novos e os alterados do `.feature` (recusa sem comprar, corrigir
  depois, tempo depois de recusar, cor repetida, posição, cortar, juntar, chat só da sala).
  "Arrumar o suporte" é `@e2e`, porque é só cliente.
- **Vitest:**
  - `rack.ts`, `selection.ts` e as novas funções de `layout.ts`;
  - `describeIssue` para cada código;
  - `melds.json` com issues.
- **playwright-bdd:** suporte, mover várias peças, chat e balão.

### 9.7 Riscos

- Com quatro jogadores e um suporte de duas fileiras, a tela fica cheia. O alvo é 1280×720.
  Abaixo disso, a mesa e o suporte diminuem juntos (escala CSS).
- `rummy.move` fica maior por causa das posições: são até 106 peças e cerca de 40 combinações,
  menos de 2 KB. O limite de leitura do ws continua alto o bastante (sobe para 8 KB).
- Mudança de protocolo: `rummy.move`, `rummy.state` e `rummy.draft` passam a usar
  `melds: [{tiles, x, y}]`. Servidor e cliente estão na mesma versão do repositório, então não
  há compatibilidade a manter.

### 9.8 O que mudou durante a implementação da emenda

- Visual da mesa no estilo da folha de assets do usuário: botões, plaquinhas, aro dos avatares,
  mesa oval e monte desenhados em CSS (recortar a folha ficava serrilhado ao ampliar).
- Os steps playwright dos cenários `@ui` e `@e2e` (T18, T31) ficaram para depois do merge; o
  teste no navegador foi manual.

| Mudança | Task | Por quê |
|---|---|---|
| As posições vão num array paralelo, `pos: [[x, y], ...]`, e não dentro de cada combinação | T23 | Menos mudanças no cliente e no servidor; `melds` continua igual |
| Um `rummy.move` sem `pos` recebe posições automáticas (`AutoPos`, em linhas de três) | T23 | Mantém os testes antigos e os steps simples |
| O chat da mesa usa `rummy.say` (cliente) e `rummy.chat` (servidor), com o mesmo limitador do chat da Praça | T24 | Um só caso de uso, `SendChat`, com `Table: true` |
| Na partida, o chat da Praça some e o chat da mesa fica no lugar | T30 | A cena ocupa a tela inteira |
| Cortar, juntar, mover várias peças e arrumar o suporte viraram cenários `@ui` | T20 | São gestos do navegador; o servidor só recebe a mesa resultante |
| Em telas com menos de 1100 px, o chat começa fechado, no canto de cima | T30 | Ele cobria os botões |

### 9.9 Tasks (vão para `tasks.md`)

- T20: cenários alterados e novos no godog (RED).
- T21: `Issue` em `Classify` e na abertura, mais `melds.json`.
- T22: recusa sem encerrar a vez, com `Refused` e `rummy.refused`.
- T23: posições em `Meld` e no protocolo.
- T24: chat da sala.
- T25: `melds.ts` com issues e `describeIssue`.
- T26: `layout.ts`: posições, `cutAt`, `joinEnds` e `moveMany`.
- T27: `rack.ts` e `selection.ts`.
- T28: `RummyTableScene`: feltro, lugares, anel e monte.
- T29: `TileRack`, arrasto e seleção.
- T30: `TableChat` e balões.
- T31: teste no navegador e playwright.
