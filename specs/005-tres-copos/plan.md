# RW-005 — Plano técnico: Minigame dos três copos

Status: **Aprovado** (2026-10-26) · Spec: [spec.md](spec.md) · Criado: 2026-10-26

## 1. Visão geral

```
Navegador                                   Servidor Go
┌───────────────────────────────┐   WS   ┌───────────────────────────────────────┐
│ aviso "Aperte E" (perto)      │ ─────► │ adapter/ws: cups.join / pick / leave  │
│ painel dos copos (React/DOM)  │        │        │                              │
│ copos na barraca (PixiJS)     │ ◄───── │ usecase: JoinCups, PickCup, LeaveCups │
│ timeline: onde está cada copo │ eventos│        │                              │
└───────────────────────────────┘        │ domain/cups: a partida (máquina de    │
                                         │ estados) · domain/room: barraca, trava│
                                         │ RoomRunner: avança a partida no tick  │
                                         └───────────────────────────────────────┘
```

- **A partida vive no servidor**, dentro da sala, no mesmo goroutine do `RoomRunner`. Ela é
  avançada pelo tick de 20 Hz que já existe, então não precisa de locks nem de timers próprios.
- O navegador só **anima** o que o servidor mandou e **envia intenções** (entrar, escolher, sair).
- Nenhuma dependência nova.

## 2. Barraca no mapa (RC1, RC2)

- O `praca.json` ganha a barraca na área livre de cima, entre as árvores e a fonte:
  tiles `B` nas colunas 13–15 da linha 4 (3×1 tiles, 96×32 px). `B` é **sólido** como `#`.
- O JSON ganha `"booth": { "x": 448, "y": 128, "w": 96, "h": 32 }`, em pixels.
  É a fonte única: o servidor usa para validar "perto", e o cliente para desenhar e mostrar o aviso.
- **Perto** = os pés do guaxinim a no máximo **48 px** do retângulo da barraca
  (`BoothReach`). A mesma função existe no Go (`TileMap.NearBooth`) e no TypeScript
  (`nearBooth`), com a mesma tabela de casos nos dois testes.
- Os cenários da RW-001 que usam obstáculos (fonte e árvores) não mudam, porque a barraca fica
  longe deles.

## 3. Backend

### 3.1 Domínio: `domain/cups` (novo pacote, sem imports externos)

A partida é uma **máquina de estados pura**, avançada por `Advance(now)`. Ela não lê o
relógio nem sorteia sozinha: recebe `now` e um `Shuffler` (interface), então os testes
controlam o tempo e o sorteio.

```
           E (1º jogador)            fim da espera              5ª rodada revelada
  Idle ─────────────────► Lobby ─────────────────► Round ──────────────────────► Results
   ▲                     (10 s)                    │  ▲                            │
   │                                               ▼  │ (próxima rodada)           │
   │                                    Show → Shuffle → Choose → Reveal ──────────┘
   │          todos saíram (de qualquer fase)                                      │
   └───────────────────────────────────────────────────────────────────────────────┤
                       fim do placar: inscritos para a próxima? → Lobby, senão Idle
```

| Conceito | Regra |
|---|---|
| `Game` | Fase, jogadores da partida, inscritos na próxima, rodada atual, placar |
| Slots | Os copos ficam em 3 posições (0, 1, 2 = 1, 2, 3 da esquerda para a direita). A noz começa num slot; uma **troca** `[a, b]` troca o conteúdo de dois slots. A escolha é um slot |
| `Shuffler` | `Start() int` (slot da noz) e `Swaps(n int) [][2]int`. Em produção usa `math/rand/v2` (não precisa ser criptográfico); nos testes, sequências fixas |
| `Join(id, now)` | Idle → abre a espera (RC4); Lobby → entra; outra fase → inscrito na próxima (RC5). Repetido é ignorado |
| `Pick(id, slot, now)` | Só na fase Choose, só de quem joga, só uma vez (RC9, RC18). Retorna erro tipado, que o caso de uso ignora em silêncio |
| `Leave(id)` | Sai da partida e do placar (RC13). Sem ninguém → Idle (RC14). Se todos que restam já escolheram, a rodada revela na hora (RC15) |
| `Advance(now)` | Muda de fase pelos prazos e devolve os **eventos** que aconteceram |
| Placar | Ordenado por acertos (desc) e, no empate, pela ordem de entrada. A posição é igual no empate (RC12) |

**Constantes** (Q4), num só arquivo:

| Rodada | Trocas | Duração de cada troca |
|---|---|---|
| 1 | 3 | 600 ms |
| 2 | 5 | 500 ms |
| 3 | 6 | 400 ms |
| 4 | 8 | 320 ms |
| 5 | 10 | 250 ms |

Espera: 10 s · mostrar a noz: 1,5 s · escolher: 10 s · revelação: 2,5 s · placar: 6 s.

### 3.2 Sala: travar o guaxinim (RC21)

- `Room` ganha `SetBusy(id, bool)`. O `Step` não move quem está ocupado, e um input recebido
  nesse estado é guardado mas não move. Ao sair da partida, o input é zerado, para o guaxinim
  não sair andando sozinho.
- O caso de uso marca ocupado ao entrar na espera ou na partida, e libera ao sair ou no fim.

### 3.3 Casos de uso

- `JoinCups`, `PickCup` e `LeaveCups`, no mesmo estilo dos atuais (`Exec` no `RoomRunner`).
  - `JoinCups` confere `NearBooth` com a posição do servidor. Longe → evento `CupsRejected{too_far}`
    só para o jogador (RC2).
- O `RoomRunner.step()` chama `game.Advance(clock.Now())` em todo tick e traduz os eventos
  do domínio em eventos `usecase` para o `Broadcaster`.
- `LeaveRoom` e a expiração por silêncio (RN5) também chamam `game.Leave` (RC15).
- O `Joined` (entrada na sala) passa a levar o estado atual da barraca, para quem chega no meio
  de uma partida ver os copos e os sinais certos.

### 3.4 Protocolo WebSocket (v1, só adições)

Mensagens novas não quebram clientes antigos (tipos desconhecidos já são ignorados),
então o `v` continua 1 (constituição, VI).

**Cliente → servidor**

| type | data |
|---|---|
| `cups.join` | `{}` |
| `cups.pick` | `{ "round": 1, "slot": 0 }` (`round` evita que uma escolha atrasada caia na rodada seguinte) |
| `cups.leave` | `{}` |

**Servidor → cliente**

| type | para | data |
|---|---|---|
| `cups.status` | sala | `{ "phase": "idle"\|"lobby"\|"playing"\|"results", "players": [id], "queued": [id], "lobbyMs": int }` — quem joga (sinal sobre o guaxinim, RC22), inscritos e tempo da espera (RC6). Enviado a cada mudança |
| `cups.rejected` | jogador | `{ "reason": "too_far" }` |
| `cups.round` | sala | `{ "round": 1, "total": 5, "nut": 0, "showMs": 1500, "swaps": [[0,1],[1,2]], "swapMs": 600, "pickMs": 10000 }` |
| `cups.picked` | partida | `{ "id": "p1" }` — alguém escolheu, **sem o copo** (RC10) |
| `cups.reveal` | sala | `{ "round": 1, "nut": 2, "picks": [{ "id": "p1", "slot": 2, "hit": true }] }` (`slot` é `null` para quem não escolheu) |
| `cups.results` | sala | `{ "scores": [{ "id": "p1", "hits": 4, "place": 1 }] }` |

`room.joined` ganha o campo `cups` com o mesmo formato de `cups.status`, mais a rodada em
andamento (se houver), com o tempo já decorrido.

**Por que mandar todas as trocas de uma vez:** a animação precisa delas, e quem as vê é
justamente quem está olhando os copos. Um cliente "esperto" que lê as trocas só descobre o que
um jogador atento também descobre (RC17). O servidor nunca manda as escolhas antes da revelação.

## 4. Frontend

### 4.1 Lógica pura (`src/game/cups/`, testada com Vitest)

- `timeline.ts`: dado `cups.round` e o tempo desde que ele chegou, devolve a fase
  (`show`/`shuffle`/`choose`) e a posição x de cada copo (interpolada durante a troca, com um
  arco para o copo da frente e o de trás não se cruzarem). É a mesma função para o painel e
  para a barraca, então os dois ficam sincronizados (RC20).
- `booth.ts`: `nearBooth` (mesma regra do Go).
- `cupsState.ts`: reducer das mensagens `cups.*`, no estilo do `roomState.ts`.

### 4.2 Painel (RC19): React + DOM

- `CupsPanel.tsx` abre por cima da Praça para quem está na espera ou na partida.
  - Espera: lista de inscritos e contagem.
  - Rodada: os três copos e a noz em CSS (formas simples, Q2), animados pela `timeline`
    com `transform`. Cada copo é um `<button>` com `aria-label="Copo 1"` e `data-slot`.
  - Resultado da rodada, placar final, **Jogar de novo** e **Sair**.
- Teclas 1, 2, 3 escolhem e Esc sai, só quando o chat não está em foco (como o WASD).
- **Por que DOM e não PixiJS:** os copos viram botões de verdade (clique, teclado,
  acessibilidade), e os testes de navegador leem o estado do DOM, como já fazem com a lista
  acessível. A alternativa da tela própria (RC19) fica guardada e não é implementada.

### 4.3 Na Praça (PixiJS)

- A barraca e três copos pequenos desenhados com `Graphics` na `scene.ts`, ordenados por
  profundidade como os outros obstáculos. Os copos seguem a mesma `timeline` (RC20).
- Aviso "Aperte E para jogar" acima da barraca quando `nearBooth` é verdadeiro.
- Sinal de jogando: um copinho ao lado do nome de quem está em `players` (RC22).
- A lista acessível de jogadores ganha `data-cups="playing"|"queued"`.

## 5. Testes

| Camada | Ferramenta | Como |
|---|---|---|
| Cenários da spec | **godog** | Steps chamam os casos de uso com `FakeClock`, `FakeBroadcaster` e um `Shuffler` fixo. "Passam 10 segundos" avança o relógio e dá ticks. Todos os 22 cenários rodam aqui |
| Smoke | godog + servidor real | `cups.join` e uma rodada completa por WebSocket de verdade (`@smoke`) |
| Navegador | **playwright-bdd** | `@e2e`: aviso perto da barraca, entrar, acertar e errar, copos na barraca para quem não joga, sinal de jogando, guaxinim parado. O step descobre o copo certo lendo as mensagens `cups.round` do WebSocket (`page.on("websocket")`), como um jogador atento faria |
| Domínio | `go test` | Máquina de estados, trocas, placar com empate, saída no meio |
| Frontend | Vitest | `timeline`, `nearBooth`, reducer, protocolo (mesmos JSONs dos testes Go) |

Para os testes de navegador não levarem minutos, o servidor aceita
`RW_CUPS_FAST=1`, que divide todos os tempos por 10 (só em teste; o padrão é o tempo real).

## 6. Riscos

- **Relógios diferentes:** o cliente anima a partir da chegada de `cups.round`. Com atraso de
  rede, a animação termina um pouco depois para ele; a fase de escolha (10 s) absorve isso, e o
  servidor é quem decide o prazo.
- **Abas em segundo plano** pausam a animação. Ao voltar, a `timeline` calcula a posição pelo
  tempo, então os copos pulam para o lugar certo em vez de ficarem atrasados.
- **Testes de navegador mais lentos:** mitigado pelo `RW_CUPS_FAST`.

## 7. Decisões deste plano (aprovadas em 2026-10-26)

1. **Partida dentro da sala**, no tick que já existe, em um pacote de domínio novo `cups`.
2. **Barraca** em cima da fonte (colunas 13–15, linha 4), com "perto" = até 48 px.
3. **Trocas enviadas de uma vez** no começo da rodada (seção 3.4).
4. **Painel em React/DOM**; a barraca na Praça em PixiJS.
5. **`RW_CUPS_FAST`** para acelerar os tempos só nos testes de navegador.
