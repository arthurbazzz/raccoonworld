# RW-001 — Sala inicial com jogadores em tempo real

Status: **Aprovada** (2026-10-25) · Criada: 2026-10-25

## Problema

O Raccoons World precisa de um primeiro "mundo" jogável que prove o núcleo do produto:
vários jogadores no mesmo lugar, vendo uns aos outros se moverem e conversarem em
tempo real. Sem isso não existe interação social, e os minigames não têm onde existir.

## Objetivo

Um jogador abre o jogo, escolhe um apelido, entra na sala inicial (a "Praça") e:
- vê o próprio guaxinim e os dos outros jogadores presentes;
- anda pela sala com **WASD**, e todos veem o movimento;
- envia mensagens curtas que aparecem como balão sobre o guaxinim;
- vê os outros jogadores entrando e saindo.

## Fora de escopo

- Autenticação e contas (adiada). Nesta spec o jogador é um **visitante** com apelido.
- Persistência de progresso, inventário, moeda e customização do avatar.
- Mais de uma sala, e os minigames.
- Moderação avançada de chat. Nesta spec o chat é **livre**, só com limite de tamanho e
  de frequência. O chat restrito (filtro ou frases prontas) terá uma spec própria.
- Movimento por clique e controles de celular.

## Usuários

- **Visitante**: qualquer pessoa que abre o jogo no navegador.

## Regras de negócio

- RN1. O apelido tem de 3 a 16 caracteres (letras, números e `_`) e é único **entre
  os jogadores conectados** à sala. A unicidade **ignora maiúsculas e minúsculas**:
  com "rocky" na sala, "Rocky" e "ROCKY" são recusados. O apelido é exibido como o
  jogador digitou.
- RN2. A Praça tem capacidade máxima de **10** jogadores.
- RN3. O movimento é feito com WASD: W para cima, A para a esquerda, S para baixo e D para a direita,
  e as diagonais combinam teclas. O cliente envia só as teclas pressionadas. O servidor
  calcula a posição com velocidade fixa e impede atravessar obstáculos ou sair da sala.
- RN3.1. Enquanto o campo de chat está em foco, WASD digita texto e não move o guaxinim.
- RN6. Visual em **perspectiva 3/4**, uma mistura de top-down e isométrico: o chão é visto de cima
  e os personagens e objetos aparecem levemente de frente, com profundidade pela posição vertical
  (quem está mais abaixo é desenhado na frente).
- RN4. Mensagens de chat têm no máximo 120 caracteres, com no máximo 1 mensagem por segundo por jogador.
  - O limite conta **caracteres**, não bytes ("é" vale 1).
  - Mensagens vazias ou só com espaços são descartadas.
  - Mensagens descartadas não contam para o limite de 1 por segundo.
- RN5. Se a conexão cair, o guaxinim some para os outros em até 10 segundos.

## Critérios de aceitação

Os cenários executáveis ficam em [`features/`](features/). Na T3 entraram três
cenários que detalham regras desta spec:
- **Apelido inválido** virou um Esquema do Cenário com 5 exemplos (RN1).
- **Cada tecla move em uma direção** cobre W, A, S e D separadamente (RN3).
- **Mensagem longa demais** cobre o limite de 120 caracteres (RN4).

O cenário "Digitar no chat não move o guaxinim" tem a tag `@ui`, porque só existe
no navegador. Ele roda no playwright-bdd e o godog o ignora.

Os cenários com a tag `@smoke` ("Visitante entra na Praça", "Apelido já em uso" e
"Mensagem de chat") também rodam contra o **servidor real**, por WebSocket de verdade.

Os cenários com `@ui` ou `@e2e` ("Visitante entra na Praça", "Andar com WASD", "Mensagem de chat"
e "Digitar no chat não move o guaxinim") rodam também **no navegador**, com o playwright-bdd, um
contexto de navegador por jogador e os dois servidores reais. O balão é desenhado no canvas, então
o teste confere a mesma mensagem na lista acessível de mensagens.

```gherkin
# language: pt
Funcionalidade: Sala inicial em tempo real

  Contexto:
    Dado que a sala "Praça" está aberta

  Cenário: Visitante entra na Praça
    Quando o visitante entra com o apelido "rocky"
    Então "rocky" aparece na Praça
    E "rocky" recebe a lista dos jogadores já presentes

  Cenário: Apelido já em uso
    Dado que "rocky" está na Praça
    Quando outro visitante tenta entrar com o apelido "rocky"
    Então a entrada é recusada com o motivo "apelido em uso"

  Cenário: Apelido inválido
    Quando o visitante tenta entrar com o apelido "a"
    Então a entrada é recusada com o motivo "apelido inválido"

  Cenário: Sala cheia
    Dado que a Praça está com a capacidade máxima
    Quando o visitante tenta entrar com o apelido "bandit"
    Então a entrada é recusada com o motivo "sala cheia"

  Cenário: Outros jogadores veem a chegada
    Dado que "rocky" está na Praça
    Quando "bandit" entra na Praça
    Então "rocky" é avisado de que "bandit" entrou

  Cenário: Andar com WASD
    Dado que "rocky" e "bandit" estão na Praça
    Quando "rocky" segura a tecla "D"
    Então "rocky" se move para a direita
    E "bandit" vê "rocky" se mover para a direita

  Cenário: Andar na diagonal
    Dado que "rocky" está na Praça
    Quando "rocky" segura as teclas "W" e "D"
    Então "rocky" se move para cima e para a direita
    E a velocidade de "rocky" é a mesma do movimento em linha reta

  Cenário: Parar ao soltar a tecla
    Dado que "rocky" está segurando a tecla "D"
    Quando "rocky" solta a tecla "D"
    Então "rocky" para de se mover

  Cenário: Obstáculo bloqueia o movimento
    Dado que "rocky" está encostado em um obstáculo à sua direita
    Quando "rocky" segura a tecla "D"
    Então a posição de "rocky" não atravessa o obstáculo

  Cenário: Limite da sala
    Dado que "rocky" está na borda esquerda da Praça
    Quando "rocky" segura a tecla "A"
    Então "rocky" continua dentro da Praça

  Cenário: Digitar no chat não move o guaxinim
    Dado que "rocky" está com o campo de chat em foco
    Quando "rocky" digita "wasd"
    Então a posição de "rocky" não muda

  Cenário: Mensagem de chat
    Dado que "rocky" e "bandit" estão na Praça
    Quando "rocky" envia a mensagem "olá!"
    Então "bandit" vê o balão "olá!" sobre "rocky"

  Cenário: Mensagens rápidas demais
    Dado que "rocky" acabou de enviar uma mensagem
    Quando "rocky" envia outra mensagem em menos de 1 segundo
    Então a segunda mensagem é descartada

  Cenário: Jogador desconecta
    Dado que "rocky" e "bandit" estão na Praça
    Quando a conexão de "rocky" cai
    Então em até 10 segundos "bandit" deixa de ver "rocky"
```

## Decisões

- **Q1. Capacidade:** 10 jogadores (RN2).
- **Q2. Visual:** perspectiva 3/4, uma mistura de top-down e isométrico (RN6).
- **Q3. Movimento:** WASD (RN3).
- **Q4. Chat:** livre por enquanto. O chat restrito fica para uma spec futura.
