# RW-006 — Rummy das Nozes (jogo de peças)

Status: **Aprovada** (2026-10-30) · Emenda A: **Aprovada** (2026-10-30) · Criada: 2026-10-30 · Depende de: RW-001, RW-005

## Problema

A Praça tem um minigame rápido (RW-005), mas nenhum jogo de **estratégia por turnos** em que
os jogadores sentem à mesa juntos por alguns minutos. Um jogo de peças estilo rummy é conhecido,
funciona bem com 2 a 4 pessoas e dá motivo para conversar no chat enquanto se joga.

## Objetivo

Na Praça existe uma **mesa de peças**. Nela os jogadores abrem **salas**, e em cada sala de
2 a 4 jogadores jogam uma partida, cada um
recebe uma mão de peças escondida e, na sua vez, baixa combinações na mesa ou compra uma peça.
Ganha quem **ficar sem peças** primeiro.

## Fora de escopo

- Moeda, prêmios, ranking e placar salvo (como na RW-005).
- Mais de uma mesa física na Praça, e mesas em outras salas do mundo. Várias partidas
  acontecem na mesma mesa, em salas separadas (RN5).
- Adversários controlados pelo computador.
- Controles de celular e de toque.
- Efeitos sonoros e música.
- Assistir uma partida de dentro do painel: quem não joga só vê a mesa no mundo (RN25).

## Usuários

- **Visitante** na Praça (RW-001), com o seu guaxinim.

## As peças

- RN1. São **106 peças**: números de **1 a 13** em **4 cores**, duas cópias de cada, mais
  **2 coringas**. As cores e o desenho seguem o tema dos guaxinins (Q6).
- RN2. Uma **combinação** tem pelo menos 3 peças e é de um destes tipos:
  - **Grupo:** o mesmo número em cores diferentes (3 ou 4 peças);
  - **Sequência:** a mesma cor em números seguidos (ex.: 4-5-6-7). O 1 não vem depois do 13.
- RN3. O **coringa** vale por qualquer peça que complete a combinação.

## Entrar na partida

- RN4. Ao chegar perto da mesa, o jogador vê "Aperte **E** para jogar". Longe dela, o aviso
  some e a tecla E não faz nada.
- RN5. **Salas:** ao apertar E, o jogador vê a lista de salas da mesa, com o modo, os
  jogadores e as vagas de cada uma, e pode:
  - **entrar** numa sala em espera com vaga, ou
  - **criar uma sala**, escolhendo o **modo** (RN13). Ele fica sozinho na espera.
  Cada sala tem **até 4 jogadores** e uma partida própria. Várias salas jogam ao mesmo tempo,
  sem limite além da capacidade da Praça (10). Uma sala some quando todos saem dela.
- RN6. Com **2 ou mais** inscritos, começa uma contagem de **15 segundos** (Q3). Com 4 inscritos
  a partida começa na hora. Se os inscritos voltarem a ser só 1, a contagem para.
- RN7. Salas com partida em andamento aparecem na lista como "jogando" e não aceitam
  entradas. No fim da partida, quem escolhe **Jogar de novo** volta para a espera da mesma sala.

## A partida

- RN8. Cada jogador recebe **14 peças**. O resto forma o **monte**, virado para baixo.
  A ordem dos turnos é sorteada.
- RN9. A **mão é secreta**: cada jogador vê só as suas peças. Todos veem a mesa, quantas peças
  cada um tem na mão e quantas restam no monte. Ninguém vê a mesa nem as mãos de outras salas.
- RN10. Na sua vez, o jogador:
  - **baixa** peças da mão formando combinações novas ou aumentando as da mesa, **ou**
  - **compra** uma peça do monte, o que encerra a vez.
- RN11. **Mexer na mesa:** depois de abrir (RN13), o jogador pode desmontar e remontar
  qualquer combinação da mesa durante a vez, contanto que, ao terminar, **todas** as combinações
  sejam válidas e nenhuma peça da mesa vá para a mão.
- RN12. Durante a vez, as combinações podem ficar inválidas. O jogo destaca as inválidas, e o
  botão **Desfazer** devolve a mesa e a mão como estavam no começo da vez.
- RN12a. **Erros claros (Q11):** cada erro aponta **a combinação e as peças culpadas**, que
  ficam destacadas na mesa, e explica em linguagem simples o que está errado e como resolver.
  O erro aparece na hora, enquanto o jogador monta a mesa, e não só quando ele termina a vez.
  Exemplos:
  - cor repetida: "O grupo de 7 tem dois 7 verdes. Tire um deles.";
  - poucas peças: "Esta combinação tem só 2 peças; precisa de pelo menos 3.";
  - cores misturadas: "Sequência com cores diferentes: o 5 vermelho não combina com as azuis.";
  - número faltando: "Falta o 6 azul entre o 5 e o 7.";
  - fora da faixa: "A sequência passou do 13.";
  - grupo grande demais: "Um grupo tem no máximo 4 peças, uma de cada cor.";
  - abertura baixa: "Sua abertura soma 24 pontos; faltam 6 para 30.";
  - abertura com peça da mesa: "Na abertura use só peças do seu suporte; o 4 azul é da mesa.";
  - nada baixado: "Você não baixou nenhuma peça. Baixe uma combinação ou clique em Comprar."
- RN13. **Modos** (Q2), escolhidos ao abrir a espera:
  - **Clássico:** a primeira jogada de cada jogador (a **abertura**) precisa usar só peças da
    mão e somar pelo menos **30 pontos** (soma dos números; o coringa vale o número que
    substitui). Antes de abrir, o jogador não mexe nas combinações da mesa.
  - **Livre:** não há abertura mínima. Desde a primeira vez, o jogador pode baixar e mexer na
    mesa.
- RN14. Ao clicar **Terminar vez**, o servidor confere tudo (RN2, RN11, RN13). **Se a jogada
  for inválida, a vez não acaba** (Q12): a mesa e a mão ficam como o jogador deixou, ele vê o
  erro (RN12a) e pode corrigir, tentar outra jogada, Desfazer ou Comprar, enquanto houver
  tempo. Ele não compra peça por ter errado. Se não baixou nenhuma peça, o botão Terminar vez
  fica desabilitado e ele usa Comprar.
- RN15. **Tempo da vez:** **40 segundos**, com um relógio visível para todos. Se o tempo
  acabar, a vez é desfeita (mesa e mão voltam ao começo da vez) e o jogador **compra uma peça
  automaticamente**.
- RN16. **Monte vazio:** quem não pode ou não quer jogar passa a vez sem comprar. Se todos
  passarem uma volta inteira seguida, a partida acaba (RN18).

## Fim da partida

- RN17. Ganha quem **ficar sem peças** na mão ao terminar uma vez válida.
- RN18. Se a partida acabar com o monte vazio (RN16), ganha quem tiver **menos pontos na mão**
  (coringa conta 30). Empates aparecem empatados.
- RN19. No fim, todos veem o **placar**: o vencedor e os pontos que sobraram na mão de cada um,
  com as opções **Jogar de novo** e **Sair**.

## Sair e desconectar

- RN20. O jogador pode **sair a qualquer momento** (Esc ou botão Sair). As peças dele voltam
  ao monte, embaralhadas, e a partida continua.
- RN21. Se o jogador **desconectar**, a partida espera **30 segundos** pela volta dele (Q4).
  Se for a vez dele, o relógio da vez continua (RN15). Se voltar a tempo, ele continua com a
  mesma mão; se não, sai como em RN20.
- RN22. Se sobrar **um só** jogador, ele ganha a partida.

## Justiça (servidor autoritativo)

- RN23. O servidor embaralha, distribui, guarda o monte e as mãos e confere cada vez. O
  navegador só mostra a mesa e envia a jogada.
- RN24. Nenhum jogador recebe as peças da mão de outro nem a ordem do monte.

## Na Praça

- RN25. A mesa é um obstáculo fixo na Praça, desenhada com a arte de `web/art/rummy.webp`
  (Q6). Quem não está jogando vê, sobre a mesa, quantas salas estão jogando, mas não as peças.
- RN26. Para quem joga, um **painel** abre por cima da Praça, como na RW-005, com a mesa, a mão
  (que pode ser ordenada por cor ou por número), o relógio e os botões.
- RN27. Enquanto joga, o guaxinim fica parado e o chat continua funcionando. Os outros veem a
  etiqueta "jogando" sobre ele.

## Emenda A: mesa, suporte, chat e mesa livre (2026-10-30)

Pedida pelo dono depois do primeiro teste. Substitui RN26 e completa RN11 e RN27.

### Visual da partida

- RN28. **Mesa em tela cheia** (Q7), no estilo dos jogos de cartas online. A partida abre numa
  cena própria por cima da Praça, com uma **mesa oval de feltro** no centro.
  - Os adversários ficam ao redor (em cima, à esquerda e à direita, conforme quantos são),
    cada um com o avatar do seu guaxinim, o apelido e o número de peças.
  - O **monte** fica num canto da mesa, mostrando quantas peças restam.
- RN29. **Vez visível:** o avatar de quem joga ganha um **anel que esvazia** ao longo dos 40 s
  da vez (RN15). Nos últimos 10 s o anel fica vermelho.
- RN30. **Suporte de peças:** a mão fica embaixo, num **suporte escuro com duas fileiras**, como
  o porta-peças do jogo físico. As peças ficam em pé, com vagas livres entre elas.
  - O jogador **arruma o suporte como quiser**: arrasta uma peça para qualquer vaga das duas
    fileiras e ela fica ali, sem se reordenar sozinha.
  - Os botões **Por cor** e **Por número** reorganizam o suporte de uma vez.
  - A arrumação é só visual e local: o servidor não recebe nem guarda as posições.
  - Uma peça comprada entra na primeira vaga livre.
- RN31. Os botões **Terminar vez**, **Comprar**, **Desfazer** e **Sair** ficam ao lado do
  suporte. Fora da vez, só **Sair** aparece.

### Mesa livre (Q9)

- RN32. **Posição livre:** cada combinação tem uma posição na mesa.
  - O jogador da vez arrasta uma combinação inteira, pela alça dela, para **qualquer ponto
    livre** da mesa, e todos da sala veem a nova posição.
  - As combinações não se sobrepõem: uma combinação solta em cima de outra é empurrada para o
    lugar livre mais perto.
- RN33. **Dividir:**
  - Soltar uma peça entre duas peças de uma combinação insere a peça ali.
  - Arrastar peças **para fora** de uma combinação cria uma combinação nova onde elas forem
    soltas.
  - Clicar no espaço entre duas peças de uma combinação **corta** a combinação em duas naquele
    ponto.
- RN34. **Juntar:** arrastar uma combinação até a ponta de outra junta as duas numa só.
- RN35. **Várias peças:**
  - **Shift+clique** marca e desmarca peças, do suporte ou da mesa.
  - Um **retângulo de seleção**, arrastado num espaço vazio da mesa, marca as peças dentro
    dele.
  - Ao arrastar uma peça marcada, todas as marcadas vão juntas, na ordem em que estavam.
  - **Esc** desmarca tudo.
- RN36. As regras de mexer na mesa continuam as mesmas (RN11, RN13):
  - no modo Clássico, antes de abrir, o jogador não move nem divide as combinações da mesa;
  - peças da mesa nunca voltam para o suporte;
  - Desfazer volta tudo ao começo da vez, inclusive as posições.

### Chat da mesa (Q8)

- RN37. A partida tem um **chat da sala**, num painel recolhível ao lado da mesa. Só os
  jogadores daquela sala enviam e recebem essas mensagens. Quem está na Praça não as vê.
- RN38. Cada mensagem também aparece por 5 segundos num **balão** ao lado do avatar de quem a
  enviou.
- RN39. As mensagens seguem os limites do chat da Praça (RW-001): tamanho máximo e limite de
  envio. O histórico dura só a partida.

## Critérios de aceitação

Rascunho dos cenários. A versão final vai para `features/` na fase de tasks.

```gherkin
# language: pt
Funcionalidade: Rummy das Nozes

  Contexto:
    Dado que a sala "Praça" está aberta
    E que a mesa de peças está livre

  Cenário: Criar uma sala escolhendo o modo
    Dado que "rocky" está perto da mesa
    Quando "rocky" aperta a tecla "E" e cria uma sala no modo "Clássico"
    Então começa a espera de uma sala no modo "Clássico" com "rocky" inscrito

  Cenário: Ver e entrar numa sala
    Dado que "rocky" criou uma sala no modo "Livre"
    Quando "bandit" aperta a tecla "E" perto da mesa
    Então "bandit" vê a sala de "rocky" no modo "Livre" com 3 vagas
    E quando "bandit" entra na sala, os dois estão na mesma espera

  Cenário: A contagem começa com dois jogadores
    Dado que "rocky" abriu a espera
    Quando "bandit" entra na espera
    Então começa a contagem de 15 segundos

  Cenário: Mesa cheia começa na hora
    Dado que "rocky", "bandit" e "mel" estão na espera
    Quando "pipoca" entra na espera
    Então a partida começa para os 4 jogadores

  Cenário: Duas salas jogando ao mesmo tempo
    Dado que uma sala com 4 jogadores está jogando
    Quando "tico" e "lua" entram juntos em outra sala
    E a contagem da sala deles acaba
    Então as duas partidas acontecem ao mesmo tempo
    E "tico" não vê a mesa da outra sala

  Cenário: Sala jogando não aceita entradas
    Dado que uma sala está jogando
    Quando "tico" aperta a tecla "E" perto da mesa
    Então "tico" vê essa sala como "jogando", sem a opção de entrar

  Cenário: Distribuição
    Quando começa uma partida entre "rocky" e "bandit"
    Então cada um tem 14 peças na mão
    E o monte tem 78 peças

  Cenário: Mão secreta
    Dado que "rocky" e "bandit" estão na mesma partida
    Então "bandit" vê quantas peças "rocky" tem, mas não quais

  Cenário: Abertura com menos de 30 pontos no modo Clássico
    Dado que é a vez de "rocky", que ainda não abriu, no modo "Clássico"
    Quando "rocky" baixa a sequência azul 3-4-5 e termina a vez
    Então a jogada é recusada
    E a mesa volta ao começo da vez
    E "rocky" compra uma peça

  Cenário: Abertura válida no modo Clássico
    Dado que é a vez de "rocky", que ainda não abriu, no modo "Clássico"
    Quando "rocky" baixa o grupo de 10 vermelho, azul e preto e termina a vez
    Então as peças ficam na mesa
    E é a vez do próximo jogador

  Cenário: Primeira jogada pequena no modo Livre
    Dado que é a vez de "rocky" no modo "Livre"
    Quando "rocky" baixa a sequência azul 3-4-5 e termina a vez
    Então as peças ficam na mesa

  Cenário: Reorganizar a mesa
    Dado que a mesa tem a sequência vermelha 4-5-6-7
    E que é a vez de "rocky", que já abriu, com o 7 azul e o 7 preto na mão
    Quando "rocky" tira o 7 vermelho da sequência e forma o grupo de 7 com o azul e o preto
    E termina a vez
    Então a mesa tem a sequência vermelha 4-5-6 e o grupo de 7

  Cenário: Mesa inválida ao terminar a vez
    Dado que é a vez de "rocky"
    Quando "rocky" deixa uma combinação com 2 peças e termina a vez
    Então a jogada é recusada e "rocky" compra uma peça

  Cenário: Desfazer
    Dado que "rocky" mexeu na mesa durante a vez
    Quando "rocky" clica em "Desfazer"
    Então a mesa e a mão voltam ao começo da vez

  Cenário: Tempo esgotado
    Dado que é a vez de "rocky"
    Quando passam 40 segundos sem "rocky" terminar a vez
    Então a vez de "rocky" é desfeita e ele compra uma peça
    E é a vez do próximo jogador

  Cenário: Coringa completa a combinação
    Dado que é a vez de "rocky", que já abriu
    Quando "rocky" baixa o 8 laranja, um coringa e o 10 laranja e termina a vez
    Então as peças ficam na mesa

  Cenário: Vitória
    Dado que "rocky" tem só a sequência preta 1-2-3 na mão e já abriu
    Quando "rocky" baixa as três peças e termina a vez
    Então todos veem que "rocky" ganhou
    E veem o placar com os pontos que sobraram na mão de cada um

  Cenário: Desconectar e voltar
    Dado que "rocky" e "bandit" estão na mesma partida
    Quando a conexão de "rocky" cai e volta em 20 segundos
    Então "rocky" continua na partida com a mesma mão

  Cenário: Desconectar e não voltar
    Dado que "rocky", "bandit" e "mel" estão na mesma partida
    Quando a conexão de "rocky" cai por 30 segundos
    Então "rocky" sai da partida e as peças dele voltam ao monte
    E a partida continua para "bandit" e "mel"

  Cenário: Último jogador ganha
    Dado que "rocky" e "bandit" estão na mesma partida
    Quando "bandit" aperta a tecla "Esc"
    Então "rocky" ganha a partida
```

## Decisões

- **Q1. Nome:** "Rummy das Nozes", por enquanto (dono, 2026-10-30).
- **Q2. Dois modos** (dono, 2026-10-30): Clássico, com abertura de 30 pontos, e Livre, sem
  abertura mínima. Nos dois, ganha quem baixa todas as peças. Quem abre a espera escolhe o modo.
- **Q3. 2 a 4 jogadores por sala, contagem de 15 s, várias salas na mesma mesa** (dono,
  2026-10-30). Uma só mesa física na Praça, com lista de salas (RN5).
- **Q4. Vez de 40 s; espera de 30 s por quem desconecta** (dono, 2026-10-30). Isso é diferente
  da saída em 10 s da RW-001: nesta mesa, o lugar do jogador fica guardado por 30 s.
- **Q5. Tempo esgotado:** desfaz a vez e compra uma peça (dono, 2026-10-30).
- **Q6. Arte temática** (dono, 2026-10-30): peças com cores e símbolos de guaxinim, geradas por
  imagem. O prompt fica em [`assets-prompt.md`](assets-prompt.md) e a arte entregue, em
  `web/art/rummy.webp`. As cores são marrom (bolota), vermelho (cereja), verde (folha) e azul
  (gota). **O verso é sempre o marrom**, para todas as peças: versos de cores diferentes
  revelariam a cor das peças escondidas, então os versos coloridos da arte não são usados.
- **Q7. Visual da partida** (dono, 2026-10-30): mesa em tela cheia com os adversários ao redor,
  no estilo dos jogos de cartas online, e a mão num suporte de duas fileiras, como o do jogo
  físico.
- **Q8. Chat só na mesa** (dono, 2026-10-30): as mensagens da partida não aparecem na Praça.
- **Q9. Mesa livre** (dono, 2026-10-30): mover combinações para qualquer lugar, dividir e
  juntar combinações, e mover várias peças de uma vez.
- **Q10. Sem máximo de peças na mão** (dono, 2026-10-30): o dono retirou esse pedido.
- **Grupo com 7 verde, 7 verde, 7 azul e 7 vermelho** (dono, 2026-10-30): continua inválido,
  porque um grupo não repete cor (RN2). Para ficar mais claro, a mensagem passa a dizer qual cor
  se repete (RN12a).
- **Q11. Erros mais claros** (dono, 2026-10-30): todos os erros dizem quais peças estão erradas
  e como resolver (RN12a).
- **Q12. Jogada recusada não encerra a vez** (dono, 2026-10-30): o jogador pode corrigir
  enquanto houver tempo. A compra automática só acontece quando o tempo acaba. Isso substitui a
  parte de RN14 que fazia comprar ao errar.
