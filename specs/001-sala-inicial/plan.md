# RW-001 — Plano técnico: Sala inicial

Status: **Implementado** (2026-10-25) · Aprovado: 2026-10-25 · Spec: [spec.md](spec.md) · Criado: 2026-10-25

## 1. Visão geral

```
Navegador (Next.js + PixiJS)            Servidor Go (Gin)
┌──────────────────────────┐   WS    ┌──────────────────────────────────┐
│ captura WASD / chat      │ ──────► │ adapter/ws  (decodifica JSON)    │
│ renderiza sala 3/4       │         │        │                         │
│ interpola outros players │ ◄────── │ usecase  (Join, Input, Chat...)  │
└──────────────────────────┘ snapshot│        │                         │
                                     │ domain/room (regras, física)     │
                                     │   loop de tick 20 Hz por sala    │
                                     └──────────────────────────────────┘
```

- O servidor é autoritativo: o cliente envia apenas as **teclas pressionadas** e as **mensagens**.
- Cada sala roda numa goroutine própria, que recebe comandos por um channel. Isso evita locks
  e race conditions no estado da sala.
- **Postgres não é usado nesta spec**: o estado da sala existe só em memória e o visitante não
  é persistido. O banco entra quando houver contas ou inventário.

## 2. Estrutura do repositório (monorepo)

```
raccoons-world/
  CLAUDE.md
  specs/
  server/                         # Go
    cmd/server/main.go            # composição: config, Gin, salas
    internal/
      domain/room/                # entidades e regras puras (sem import externo)
        nickname.go  player.go  room.go  vec.go  input.go  tilemap.go  chat.go
      usecase/                    # orquestra o domínio e define as portas
        join_room.go  set_input.go  send_chat.go  leave_room.go
        ports.go                  # Broadcaster, Clock
      adapter/
        ws/                       # protocolo JSON <-> DTOs dos casos de uso
        httpapi/                  # rotas Gin: GET /health, GET /ws
      infra/
        config/  clock/           # relógio real
    maps/praca.json               # tilemap da Praça (colisão), embutido via go:embed (maps.go)
    features/                     # runner godog + steps (lê specs/*/features)
  web/                            # Next.js (TypeScript)
    src/app/page.tsx              # tela de apelido + jogo
    src/game/                     # PixiJS: render, interpolação, input
    src/net/                      # cliente WS e tipos do protocolo
    e2e/                          # playwright-bdd
```

Os arquivos `.feature` ficam em `specs/001-sala-inicial/features/` e são a fonte única.
O godog (backend) e o playwright-bdd (frontend) os leem diretamente desse lugar.

## 3. Backend

### 3.1 Domínio (`domain/room`)

| Conceito | Regra (spec) |
|---|---|
| `Nickname` | Value object: valida `^[A-Za-z0-9_]{3,16}$` (RN1) |
| `Room` | Capacidade de 10 jogadores (RN2), unicidade de apelido (RN1), lista de jogadores |
| `Player` | id, apelido, posição, input atual, horário da última mensagem |
| `Input` | Flags W/A/S/D. O vetor é normalizado para a diagonal ter a mesma velocidade (RN3) |
| `TileMap` | Grade de colisão (tiles de 32 px) e ponto de nascimento (`spawn`). Movimento resolvido eixo a eixo, para o jogador deslizar na parede |
| Hitbox | Caixa de 20×12 px nos pés do guaxinim: na visão 3/4 a cabeça pode passar na frente do cenário, mas os pés não |
| `Chat` | Até 120 caracteres e intervalo mínimo de 1 s entre mensagens (RN4) |

O método `Room.Step(dt)` avança a física. Ele é determinístico, recebe `dt` e não lê o relógio.

Constantes iniciais: velocidade de 160 px/s, tick de 20 Hz (50 ms) e sala de 30×20 tiles.

### 3.2 Casos de uso e portas

```go
type Broadcaster interface {
    ToRoom(roomID string, ev Event)
    ToPlayer(playerID string, ev Event)
}
type Clock interface { Now() time.Time }
```

`Event` é uma interface fechada: só os tipos do pacote `usecase` a implementam (ex.: `Snapshot`).

O `RoomRunner` é dono da `Room` e roda numa goroutine. Comandos (`Exec`) e ticks passam pelo
mesmo loop, então a sala não precisa de locks. Cada tick avança exatamente `TickInterval`
(50 ms), sem medir o tempo real, e isso mantém a simulação determinística. Os ticks chegam
por um channel: em produção vêm de um `time.Ticker`, e nos testes são enviados à mão.

`JoinRoom`, `SetInput`, `SendChat` e `LeaveRoom` recebem DTOs próprios e nunca `*gin.Context`.
Os erros de domínio são `ErrNicknameTaken`, `ErrNicknameInvalid` e `ErrRoomFull`.

### 3.3 Transporte

- Gin expõe `GET /maps/:name`, com o JSON do tilemap. É **a mesma fonte** que o servidor usa
  para colisão, e o cliente baixa o mapa daí para desenhar a sala, sem cópia no frontend.
  O CORS segue o `RW_ALLOWED_ORIGINS`. Entrou na T20.
- Gin expõe `GET /ws`, e o upgrade para WebSocket é feito com **`github.com/coder/websocket`**.
  O Gin não tem WebSocket nativo, e essa biblioteca é mantida e usa `context`.
- Heartbeat: o servidor envia ping a cada 5 s, e cada pong (ou qualquer mensagem) chama o
  `KeepAlive`. Quem fica 10 s sem sinal de vida (`room.PresenceTimeout`) é removido pela
  própria sala no próximo tick, e a sala recebe `PlayerLeft` (RN5). Quando o socket fecha
  normalmente, o adapter chama `LeaveRoom` na hora. Com a regra no domínio, o timeout é
  testável com o relógio falso, sem esperar 10 s reais.
- Origem permitida por config (`RW_ALLOWED_ORIGINS`), para evitar uso cross-site.
- `ws.Hub` implementa o `Broadcaster`. Cada conexão tem um buffer de 64 mensagens, e uma
  conexão lenta que enche o buffer é derrubada, para nunca travar o loop da sala. Só recebe
  as mensagens da sala quem já recebeu `Joined`.
- Mensagens inválidas são ignoradas (log de debug), e a conexão continua aberta.

### 3.4 Protocolo WebSocket (v1)

Envelope: `{"type": "<nome>", "v": 1, "data": {...}}`. Implementado em
`server/internal/adapter/ws/protocol.go`. Os testes de lá fixam o JSON exato de cada mensagem.
- Campos desconhecidos são ignorados, para que clientes mais novos possam mandar dados opcionais.
- `v` ausente ou diferente de 1, `type` desconhecido ou `data` ausente/inválido são recusados.
- Listas vazias são enviadas como `[]`, nunca `null`.

**Cliente → servidor**

| type | data |
|---|---|
| `room.join` | `{ "nickname": "rocky" }` |
| `player.input` | `{ "up": bool, "down": bool, "left": bool, "right": bool, "seq": int }` (só quando muda) |
| `chat.send` | `{ "text": "olá!" }` |

**Servidor → cliente**

| type | data |
|---|---|
| `room.joined` | `{ "you": id, "players": [{id, nickname, x, y}], "map": "praca" }` |
| `room.join_rejected` | `{ "reason": "nickname_taken" \| "nickname_invalid" \| "room_full" \| "already_joined" }` (`already_joined`: a conexão já está na sala; tem prioridade sobre os outros motivos) |
| `player.joined` | `{ id, nickname, x, y }` (vai para a sala toda, inclusive quem entrou; o cliente ignora o próprio `id`) |
| `player.left` | `{ id }` |
| `room.snapshot` | `{ "tick": int, "players": [{id, x, y}] }` (a cada tick, só se algo mudou) |
| `chat.message` | `{ "from": id, "text": "olá!" }` |

As mensagens de chat descartadas pelo limite de frequência não geram resposta, como diz a spec.

## 4. Frontend

- **Next.js (App Router)**: uma página com o formulário de apelido e o jogo como client component.
- **PixiJS v8** para renderizar em canvas. É leve, 2D e mais simples que o Phaser para o escopo atual.
- **Perspectiva 3/4** (RN6): tiles do chão desenhados de cima e sprites com âncora nos pés.
  Os tiles sólidos viram blocos (face de cima e face da frente) e entram na mesma ordenação dos
  guaxinins: o bloco usa a borda de baixo e o guaxinim usa os pés + 0,5 (`web/src/game/depth.ts`).
  Assim, um guaxinim acima da fonte aparece atrás dela.
- O PixiJS é carregado só no cliente e sob demanda (`import()` dinâmico), depois do login.
  A primeira carga da página fica em cerca de 107 kB.
- **Input**: `keydown`/`keyup` mudam o estado de WASD, e o cliente só envia `player.input`
  quando esse estado muda. Os eventos são ignorados enquanto o campo de chat tem foco (RN3.1).
- **Suavidade**: os outros jogadores são **interpolados** entre os dois últimos snapshots,
  com cerca de 100 ms de atraso (`web/src/game/interpolation.ts`).
  - Cada snapshot recebe a hora de chegada no `onMessage`, fora do React.
  - Não há extrapolação: depois do último snapshot, o guaxinim fica parado.
  - Ao sair do repouso, é inserida uma amostra parada um tick antes. Assim o guaxinim
    começa a andar de onde estava, em vez de deslizar devagar a partir de uma amostra antiga. O próprio jogador também segue o servidor nesta versão,
  sem predição local. Se o atraso incomodar, a predição local fica para uma spec futura.
- Os balões de chat ficam sobre o sprite por 5 s, numa camada acima do cenário, para nunca ficarem
  escondidos atrás de uma árvore. Uma nova mensagem do mesmo jogador substitui o balão e reinicia o tempo.
- O campo de chat fica embaixo. Enter foca o campo; dentro dele, Enter envia e devolve o teclado ao
  WASD, e Esc sai sem enviar. Mensagens vazias ou com mais de 120 caracteres nem são enviadas, e um
  contador mostra o limite. As últimas 20 mensagens ficam numa lista `aria-live` escondida, para
  leitores de tela e para os testes de UI.
- Os sprites iniciais são placeholders (um guaxinim simples), e a arte final vem depois.

## 5. Estratégia de testes (BDD)

| Camada | Ferramenta | Como |
|---|---|---|
| Cenários da spec (backend) | **godog** | Os steps chamam os **casos de uso** com `FakeClock` e `FakeBroadcaster`. Ex.: "segura D" vira `SetInput` seguido de N × `Step`. Os testes são determinísticos e rápidos |
| Smoke do transporte | godog + servidor real | 2–3 cenários (entrar e chat) via WebSocket de verdade |
| Cenários de UI | **playwright-bdd** | 2 navegadores (dois jogadores): entrar, andar, chat, digitar sem mover |
| Domínio | `go test` | Tabela de casos para `Nickname`, colisão e normalização da diagonal |
| Frontend (unidade) | `vitest` | Protocolo (os mesmos payloads dos testes Go) e cliente WS com socket falso |

O cenário "desconexão em até 10 s" usa o `FakeClock` no nível de caso de uso, sem esperar 10 s reais.

## 6. Dependências novas (justificativa)

| Dependência | Motivo |
|---|---|
| `github.com/gin-gonic/gin` | Definido na constituição |
| `github.com/coder/websocket` | WebSocket para Go (o Gin não tem) |
| `github.com/cucumber/godog` | BDD no backend (constituição) |
| `pixi.js` | Renderização 2D em canvas |
| `playwright-bdd` + `@playwright/test` | BDD no frontend (constituição) |
| `vitest` | Testes unitários do frontend (protocolo, cliente WS, lógica do jogo), sem navegador. Entrou na T18 |

## 7. Riscos

- **Latência sem predição**: com mais de 150 ms de ping, o movimento do próprio jogador vai
  parecer atrasado. A mitigação, se necessária, é predição local com reconciliação, numa spec futura.
- **Uma instância só**: as salas vivem em memória, então reiniciar o servidor desconecta todos
  os jogadores. Isso é aceitável nesta fase.

## 8. O que mudou durante a implementação

As seções acima já estão atualizadas. Esta lista resume o que **não** estava no plano aprovado:

| Mudança | Task | Por quê |
|---|---|---|
| Apelido único **sem diferenciar maiúsculas** (`Nickname.Key()`) | T5 | Decisão do dono do projeto, para impedir que alguém se passe por outro |
| Mensagem vazia descartada; descartadas não contam no limite de 1/s | T9 | A spec não definia esses casos |
| Presença (timeout de 10 s) como regra **do domínio** (`Touch`/`Expire`) + `KeepAlive` | T14 | Torna o timeout testável com relógio falso |
| `Join` com prioridade: já na sala → sala cheia → apelido em uso; motivo `already_joined` | T16 | Pedido repetido recebia "apelido em uso" |
| Montagem extraída para `internal/app` | T17 | O smoke test usa o mesmo servidor do `main` |
| `GET /maps/:name` com CORS | T20 | O cliente desenha com o mesmo mapa da colisão, sem cópia |
| Lista oculta de jogadores (com posição) e de mensagens | T20/T23/T24 | Acessibilidade e asserções dos testes de UI |
| `vitest` para testes unitários do frontend | T18 | O plano só previa BDD no navegador |
| `playwright-bdd` 9.2.1 (e não 8.x) | T24 | A 8.x não suporta o Playwright 1.64 |
| Tags `@smoke` (servidor real) e `@e2e` (navegador) nos `.feature` | T17/T24 | Escolher quais cenários rodam em cada suíte |

## 9. Decisões deste plano que precisam do seu OK

1. ~~**Monorepo** com `server/` e `web/`~~ → **aprovado** (2026-10-25).
2. **PixiJS** como renderizador → **aprovado** (2026-10-25).
3. **Sem Postgres** nesta spec → **aprovado** (2026-10-25).
4. **Sem predição local** do movimento na v1 → **aprovado** (2026-10-25).
