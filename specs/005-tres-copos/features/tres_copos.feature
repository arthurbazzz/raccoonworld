# language: pt
@RW-005
Funcionalidade: Minigame dos três copos
  Como jogador na Praça
  Quero adivinhar onde está a noz junto com os outros
  Para ter algo divertido para fazer com quem está lá

  Contexto:
    Dado que a sala "Praça" está aberta
    E que a barraca dos três copos está livre

  @e2e
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
    Dado que "rocky" está jogando os três copos
    Quando "rocky" acompanha todas as trocas de copos
    Então a noz está no copo que as trocas indicam

  @smoke @e2e
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
    Dado que "rocky" está jogando os três copos
    Quando "rocky" joga as 5 rodadas
    Então cada rodada tem pelo menos tantas trocas quanto a anterior
    E as trocas ficam mais rápidas a cada rodada

  Cenário: Placar da partida
    Dado que "rocky" e "bandit" estão na mesma partida
    Quando a partida acaba com "rocky" acertando 4 rodadas e "bandit" acertando 3
    Então os dois veem o placar com "rocky" 4 de 5 primeiro e "bandit" 3 de 5 depois

  Cenário: Sair no meio da partida
    Dado que "rocky" e "bandit" estão na mesma partida
    Quando "rocky" aperta a tecla "Esc"
    Então "rocky" sai da partida
    E a partida continua para "bandit"

  Cenário: Último jogador sai
    Dado que "rocky" está jogando os três copos
    Quando "rocky" aperta a tecla "Esc"
    Então a barraca fica livre

  Cenário: Jogador desconecta no meio da partida
    Dado que "rocky" e "bandit" estão na mesma partida
    Quando a conexão de "rocky" cai
    Então em até 10 segundos "rocky" sai da partida
    E a partida continua para "bandit"

  @e2e
  Cenário: Quem passa vê os copos na barraca
    Dado que "rocky" está jogando os três copos
    E que "bandit" está na Praça sem jogar
    Quando os copos são embaralhados
    Então "bandit" vê os copos se mexendo na barraca

  @e2e
  Cenário: Outros veem quem está jogando
    Dado que "rocky" e "bandit" estão na Praça
    Quando "rocky" começa uma partida
    Então "bandit" vê o sinal de jogando sobre "rocky"

  @e2e
  Cenário: Guaxinim parado durante a partida
    Dado que "rocky" está jogando os três copos
    Quando "rocky" segura a tecla "D"
    Então a posição de "rocky" não muda
