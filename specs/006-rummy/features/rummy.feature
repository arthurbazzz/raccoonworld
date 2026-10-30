# language: pt
@RW-006
Funcionalidade: Rummy das Nozes
  Como jogador na Praça
  Quero jogar uma partida de peças com até mais três amigos
  Para ter um jogo de estratégia para jogar com calma

  Contexto:
    Dado que a sala "Praça" está aberta
    E que a mesa de peças está livre

  @e2e
  Cenário: Criar uma sala escolhendo o modo
    Dado que "rocky" está perto da mesa
    Quando "rocky" cria uma sala no modo "Clássico"
    Então começa a espera de uma sala no modo "Clássico" com "rocky" inscrito

  @e2e
  Cenário: Ver e entrar numa sala
    Dado que "rocky" criou uma sala no modo "Livre"
    Quando "bandit" abre a lista de salas perto da mesa
    Então "bandit" vê a sala de "rocky" no modo "Livre" com 3 vagas
    E quando "bandit" entra na sala de "rocky", os dois estão na mesma espera

  Cenário: A contagem começa com dois jogadores
    Dado que "rocky" criou uma sala no modo "Livre"
    Quando "bandit" entra na sala de "rocky"
    Então começa a contagem de 15 segundos na sala de "rocky"

  Cenário: Mesa cheia começa na hora
    Dado que "rocky", "bandit" e "mel" estão na espera da mesma sala
    Quando "pipoca" entra na sala de "rocky"
    Então a partida da sala de "rocky" começa com 4 jogadores

  @e2e
  Cenário: Duas salas jogando ao mesmo tempo
    Dado que "rocky", "bandit", "mel" e "pipoca" estão jogando na mesma sala
    Quando "tico" e "lua" começam uma partida em outra sala
    Então as duas salas estão jogando
    E "tico" não recebe a mesa da sala de "rocky"

  Cenário: Sala jogando não aceita entradas
    Dado que "rocky" e "bandit" estão jogando na mesma sala
    Quando "tico" abre a lista de salas perto da mesa
    Então "tico" vê a sala de "rocky" como "jogando"
    E "tico" não consegue entrar na sala de "rocky"

  @smoke
  Cenário: Distribuição
    Quando começa uma partida entre "rocky" e "bandit"
    Então cada um tem 14 peças na mão
    E o monte tem 78 peças

  Cenário: Mão secreta
    Dado que "rocky" e "bandit" estão jogando na mesma sala
    Então "bandit" sabe quantas peças "rocky" tem, mas não quais

  Cenário: Abertura com menos de 30 pontos no modo Clássico
    Dado que é a vez de "rocky", que ainda não abriu, no modo "Clássico"
    Quando "rocky" baixa a sequência "azul 3-4-5" e termina a vez
    Então a jogada de "rocky" é recusada
    E a mesa volta ao começo da vez
    E "rocky" compra uma peça

  Cenário: Abertura válida no modo Clássico
    Dado que é a vez de "rocky", que ainda não abriu, no modo "Clássico"
    Quando "rocky" baixa o grupo "10 vermelho, 10 azul, 10 verde" e termina a vez
    Então as peças de "rocky" ficam na mesa
    E é a vez do próximo jogador

  Cenário: Primeira jogada pequena no modo Livre
    Dado que é a vez de "rocky", que ainda não abriu, no modo "Livre"
    Quando "rocky" baixa a sequência "azul 3-4-5" e termina a vez
    Então as peças de "rocky" ficam na mesa

  @e2e
  Cenário: Reorganizar a mesa
    Dado que a mesa tem a sequência "vermelho 4-5-6-7"
    E que é a vez de "rocky", que já abriu, com "7 azul, 7 verde" na mão
    Quando "rocky" tira o "7 vermelho" da sequência e forma o grupo "7 vermelho, 7 azul, 7 verde"
    E "rocky" termina a vez
    Então a mesa tem a sequência "vermelho 4-5-6" e o grupo "7 vermelho, 7 azul, 7 verde"

  Cenário: Mesa inválida ao terminar a vez
    Dado que é a vez de "rocky", que já abriu, com "7 azul, 7 verde" na mão
    Quando "rocky" baixa o grupo "7 azul, 7 verde" e termina a vez
    Então a jogada de "rocky" é recusada
    E "rocky" compra uma peça

  @e2e
  Cenário: Desfazer
    Dado que é a vez de "rocky", que já abriu, com "7 azul, 7 verde" na mão
    E que "rocky" baixou o grupo "7 azul, 7 verde" sem terminar a vez
    Quando "rocky" desfaz a vez
    Então a mesa e a mão de "rocky" voltam ao começo da vez

  Cenário: Tempo esgotado
    Dado que é a vez de "rocky", que já abriu, com "7 azul, 7 verde" na mão
    Quando passam 40 segundos sem "rocky" terminar a vez
    Então a vez de "rocky" é desfeita e ele compra uma peça
    E é a vez do próximo jogador

  Cenário: Coringa completa a combinação
    Dado que é a vez de "rocky", que já abriu, com "8 marrom, coringa, 10 marrom" na mão
    Quando "rocky" baixa a sequência "8 marrom, coringa, 10 marrom" e termina a vez
    Então as peças de "rocky" ficam na mesa

  Cenário: Vitória
    Dado que é a vez de "rocky", que já abriu, só com "1 verde, 2 verde, 3 verde" na mão
    Quando "rocky" baixa a sequência "1 verde, 2 verde, 3 verde" e termina a vez
    Então todos na sala veem que "rocky" ganhou
    E veem o placar com os pontos que sobraram na mão de cada um

  Cenário: Desconectar e voltar
    Dado que "rocky" e "bandit" estão jogando na mesma sala
    Quando "rocky" sai da Praça e volta com o mesmo apelido em 20 segundos
    Então "rocky" continua na partida com a mesma mão

  Cenário: Desconectar e não voltar
    Dado que "rocky", "bandit" e "mel" estão jogando na mesma sala
    Quando "rocky" sai da Praça e não volta por 30 segundos
    Então "rocky" sai da partida e as peças dele voltam ao monte
    E a partida continua para "bandit" e "mel"

  Cenário: Último jogador ganha
    Dado que "rocky" e "bandit" estão jogando na mesma sala
    Quando "bandit" sai da mesa
    Então "rocky" ganha a partida
