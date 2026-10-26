# RW-005 — Minigame dos três copos

Status: **Aprovada** (2026-10-26) · Criada: 2026-10-26 · Depende de: RW-001, RW-004

## Problema

O Raccoons World tem uma sala onde os guaxinins andam e conversam, mas ainda não tem nada para
**fazer** lá. A constituição define o produto como um mundo que **hospeda minigames**, e as
referências (Club Penguin, Mundo Gaturro, Neopets) mostram que eles dão motivo para entrar, voltar
e **jogar junto**.

O jogo dos três copos é um bom primeiro minigame:
- tem regras que todo mundo já conhece;
- dá para jogar em menos de um minuto;
- funciona em grupo: todos olham os mesmos copos e cada um aposta no seu palpite;
- combina com o tema: guaxinins são espertos e gostam de esconder coisas.

## Objetivo

Na Praça existe uma **barraca dos três copos**. Vários jogadores entram na **mesma partida**:
1. Uma **noz** é colocada debaixo de um dos três copos, à vista de todos.
2. Os copos são **embaralhados**, iguais para todos.
3. **Cada jogador escolhe um copo**, sem ver a escolha dos outros.
4. Os copos sobem: cada um acerta ou erra.

A partida tem **rodadas** cada vez mais rápidas, e no fim todos veem o placar de todos.
Quem passa perto da barraca vê os copos se mexendo e pode entrar na próxima partida.

## Fora de escopo

- **Moeda, prêmios, ranking e recordes salvos.** Não há banco nem contas ainda. O placar aparece
  no fim da partida e não é guardado.
- **Apostas** de qualquer tipo.
- Um jogador **embaralhando** para os outros. Quem embaralha é sempre a barraca (o servidor).
- Mais de uma barraca, e barracas em outras salas.
- Controles de celular e de toque.
- Efeitos sonoros e música.
- Arte final: nesta spec, a barraca, os copos e a noz são **formas simples desenhadas em código**
  (Q2). Uma spec futura pode trocá-las por imagens.

## Usuários

- **Visitante** na Praça (RW-001), com o seu guaxinim.

## Regras

### A barraca
- RC1. A barraca é um **obstáculo fixo** na Praça, como os que já existem: os guaxinins não a
  atravessam. Os três copos ficam **sobre ela, no mundo**.
- RC2. Ao chegar **perto da barraca**, o jogador vê um aviso "Aperte **E** para jogar". Longe
  dela, o aviso some e a tecla E não faz nada.

### Entrar na partida
- RC3. A barraca tem **uma partida por vez**, e **vários jogadores** jogam nela juntos, até
  a capacidade da Praça (10).
- RC4. **Barraca livre:** o primeiro jogador que aperta E abre a **espera**, com uma contagem de
  **10 segundos**. Quem apertar E durante a espera também entra. Quando a contagem acaba, a
  partida começa com todos os inscritos.
- RC5. **Partida em andamento:** quem aperta E vê "Partida em andamento, você entra na próxima"
  e fica **inscrito na próxima partida**. A espera da próxima abre assim que a atual acaba.
- RC6. Durante a espera, o jogador vê quem já está inscrito e o tempo que falta.

### A partida
- RC7. A partida tem **5 rodadas**. Cada rodada tem mais trocas e trocas mais rápidas que a
  anterior.
- RC8. Em cada rodada, **para todos ao mesmo tempo**:
  1. a noz aparece debaixo de um copo, à vista, por um instante;
  2. o copo desce e os copos trocam de lugar, dois de cada vez;
  3. cada jogador escolhe um copo clicando nele ou com as teclas **1**, **2** e **3**
     (da esquerda para a direita);
  4. todos os copos sobem, mostrando onde a noz estava;
  5. cada jogador vê o próprio resultado ("Acertou!" ou "Errou!") e quem acertou na rodada.
- RC9. O jogador só pode escolher **depois** que o embaralhamento termina. Escolhas antes disso
  são ignoradas. Cada jogador escolhe **uma vez** por rodada, e não pode trocar.
- RC10. As escolhas ficam **escondidas** até os copos subirem: ninguém vê o palpite dos outros
  antes da revelação (para não copiar).
- RC11. A rodada tem **10 segundos** para escolher. Os copos sobem quando **todos escolheram**
  ou quando o tempo acaba, o que vier primeiro. Quem não escolheu, errou.
- RC12. No fim das 5 rodadas aparece o **placar da partida** com todos os jogadores, do maior
  para o menor número de acertos ("rocky 4 de 5, bandit 3 de 5"). Empates aparecem empatados.
  O jogador tem as opções **Jogar de novo** (entra na próxima partida) e **Sair**.

### Sair
- RC13. O jogador pode **sair a qualquer momento** (tecla Esc ou botão Sair). Só ele sai: a
  partida continua para os outros, e ele fica fora do placar.
- RC14. Se **todos** saírem, a partida acaba e a barraca fica livre.
- RC15. Se o jogador **desconectar**, ele sai da partida como em RC13 (dentro do prazo de 10 s
  da RW-001). Uma rodada que só esperava a escolha dele termina na hora.

### Justiça (servidor autoritativo)
- RC16. **O servidor decide tudo:** onde a noz começa, quais copos trocam, quando cada fase
  começa e quem acertou. O navegador só mostra a animação e envia a escolha.
- RC17. O embaralhamento é **aleatório e honesto**: as trocas que os jogadores veem são
  exatamente as que o servidor fez. A noz nunca muda de copo "por mágica", e quem acompanhar as
  trocas com atenção sempre acerta.
- RC18. O servidor só aceita a escolha de quem **está na partida**, no momento certo (RC9, RC11).

### Na Praça
- RC19. **Onde aparece o jogo** (Q1): para quem está jogando, um **painel** com os copos
  grandes abre por cima da Praça, que continua visível atrás. Se o painel não for viável, o plano
  pode trocar por uma tela própria (o jogador é "teleportado" para o minigame e volta à barraca
  no fim).
- RC20. **Quem não está jogando** vê os copos se mexendo **na barraca**, no mundo, sincronizados
  com a partida (Q3). Não vê o palpite de ninguém.
- RC21. Enquanto joga (e na espera), o guaxinim **fica parado** perto da barraca: WASD não o
  move e o chat continua funcionando.
- RC22. Os outros jogadores veem quem está jogando por uma etiqueta "jogando"
  ao lado do nome. (Era um copinho, trocado porque se confundia com os copos da barraca.)

## Critérios de aceitação

Rascunho dos cenários. A versão final vai para `features/` na fase de tasks.

```gherkin
# language: pt
Funcionalidade: Minigame dos três copos

  Contexto:
    Dado que a sala "Praça" está aberta
    E que a barraca dos três copos está livre

  Cenário: Aviso perto da barraca
    Dado que "rocky" está na Praça longe da barraca
    Quando "rocky" anda até perto da barraca
    Então "rocky" vê o aviso "Aperte E para jogar"

  Cenário: Tecla E longe da barraca não faz nada
    Dado que "rocky" está na Praça longe da barraca
    Quando "rocky" aperta a tecla "E"
    Então "rocky" não entra em nenhuma partida

  Cenário: Abrir a espera
    Dado que "rocky" está perto da barraca
    Quando "rocky" aperta a tecla "E"
    Então começa a espera de 10 segundos com "rocky" inscrito

  Cenário: Entrar durante a espera
    Dado que "rocky" abriu a espera na barraca
    E que "bandit" está perto da barraca
    Quando "bandit" aperta a tecla "E"
    Então "rocky" e "bandit" veem os dois inscritos

  Cenário: A partida começa para todos os inscritos
    Dado que "rocky" e "bandit" estão inscritos na espera
    Quando a contagem da espera acaba
    Então começa a rodada 1 de 5 para "rocky" e "bandit"
    E os dois veem a noz debaixo do mesmo copo

  Cenário: Chegar com a partida em andamento
    Dado que "rocky" está jogando os três copos
    E que "bandit" está perto da barraca
    Quando "bandit" aperta a tecla "E"
    Então "bandit" vê o aviso "Partida em andamento, você entra na próxima"
    E quando a partida de "rocky" acaba, "bandit" está inscrito na próxima espera

  Cenário: Todos veem o mesmo embaralhamento
    Dado que "rocky" e "bandit" estão na mesma partida
    Quando os copos são embaralhados
    Então "rocky" e "bandit" veem as mesmas trocas

  Cenário: O embaralhamento é honesto
    Dado que "rocky" está numa rodada
    Quando "rocky" acompanha todas as trocas de copos
    Então a noz está no copo que as trocas indicam

  Cenário: Cada um escolhe o seu copo
    Dado que "rocky" e "bandit" estão numa rodada com o embaralhamento terminado
    Quando "rocky" escolhe o copo onde está a noz
    E "bandit" escolhe um copo sem a noz
    Então "rocky" vê "Acertou!"
    E "bandit" vê "Errou!"
    E os dois veem que só "rocky" acertou a rodada

  Cenário: Escolhas escondidas até a revelação
    Dado que "rocky" e "bandit" estão numa rodada com o embaralhamento terminado
    Quando "rocky" escolhe um copo
    Então "bandit" não vê qual copo "rocky" escolheu

  Cenário: Copos sobem quando todos escolheram
    Dado que "rocky" e "bandit" estão numa rodada com o embaralhamento terminado
    Quando "rocky" e "bandit" escolhem um copo
    Então os copos sobem sem esperar os 10 segundos

  Cenário: Escolher antes do fim do embaralhamento
    Dado que "rocky" está numa rodada com os copos ainda trocando
    Quando "rocky" escolhe um copo
    Então a escolha é ignorada

  Cenário: Não dá para trocar de copo
    Dado que "rocky" e "bandit" estão numa rodada com o embaralhamento terminado
    E que "rocky" escolheu o copo 1
    Quando "rocky" escolhe o copo 2
    Então a escolha de "rocky" continua sendo o copo 1

  Cenário: Tempo esgotado
    Dado que "rocky" está numa rodada com o embaralhamento terminado
    Quando passam 10 segundos sem escolha de "rocky"
    Então a rodada conta como erro para "rocky"

  Cenário: Rodadas mais difíceis
    Quando "rocky" joga as 5 rodadas
    Então cada rodada tem pelo menos tantas trocas quanto a anterior
    E as trocas ficam mais rápidas a cada rodada

  Cenário: Placar da partida
    Dado que "rocky" e "bandit" estão na rodada 5
    E que "rocky" tem 3 acertos e "bandit" tem 2
    Quando "rocky" e "bandit" acertam a rodada 5
    Então os dois veem o placar com "rocky" 4 de 5 primeiro e "bandit" 3 de 5 depois
    E os dois veem as opções "Jogar de novo" e "Sair"

  Cenário: Sair no meio da partida
    Dado que "rocky" e "bandit" estão na mesma partida
    Quando "rocky" aperta a tecla "Esc"
    Então "rocky" sai da partida
    E a partida continua para "bandit"

  Cenário: Último jogador sai
    Dado que "rocky" está jogando os três copos sozinho
    Quando "rocky" aperta a tecla "Esc"
    Então a barraca fica livre

  Cenário: Jogador desconecta no meio da partida
    Dado que "rocky" e "bandit" estão na mesma partida
    Quando a conexão de "rocky" cai
    Então em até 10 segundos "rocky" sai da partida
    E a partida continua para "bandit"

  Cenário: Quem passa vê os copos na barraca
    Dado que "rocky" está jogando os três copos
    E que "bandit" está na Praça sem jogar
    Quando os copos são embaralhados
    Então "bandit" vê os copos se mexendo na barraca

  Cenário: Outros veem quem está jogando
    Dado que "rocky" e "bandit" estão na Praça
    Quando "rocky" começa uma partida
    Então "bandit" vê o sinal de jogando sobre "rocky"

  Cenário: Guaxinim parado durante a partida
    Dado que "rocky" está jogando os três copos
    Quando "rocky" segura a tecla "D"
    Então a posição de "rocky" não muda
```

## Decisões

- **Q1. Painel por cima da Praça**, com a tela própria ("teleporte") como alternativa se o
  painel não for viável (RC19).
- **Q2. Formas simples em código** para a barraca, os copos e a noz.
- **Q3. Os copos se mexem na barraca**, no mundo, para todos (RC20).
- **Q4. Dificuldade:** 5 rodadas, de **3 trocas lentas** na primeira até **~10 trocas rápidas**
  na última. Os números exatos ficam como constantes no plano.
- **Q5. Vários jogadores na mesma partida** (RC3–RC6), com espera de 10 s antes de começar e
  inscrição na próxima partida para quem chega no meio.
