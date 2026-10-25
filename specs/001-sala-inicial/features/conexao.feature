# language: pt
@RW-001
Funcionalidade: Conexão com a Praça
  Como jogador
  Quero que a Praça mostre só quem está realmente conectado
  Para não ver guaxinins "fantasmas"

  Contexto:
    Dado que a sala "Praça" está aberta

  Cenário: Jogador desconecta
    Dado que "rocky" e "bandit" estão na Praça
    Quando a conexão de "rocky" cai
    Então em até 10 segundos "bandit" deixa de ver "rocky"
