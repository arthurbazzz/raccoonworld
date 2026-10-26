# language: pt
@RW-002 @ui
Funcionalidade: Visual do guaxinim
  Como jogador
  Quero que os guaxinins olhem para onde andam e se mexam ao andar
  Para a Praça parecer viva e eu achar o meu guaxinim

  Contexto:
    Dado que a sala "Praça" está aberta

  Esquema do Cenário: Olhar para a direção do movimento
    Dado que "rocky" e "bandit" estão na Praça
    Quando "rocky" segura a tecla "<tecla>"
    Então "bandit" vê "rocky" olhando para <direção>

    Exemplos:
      | tecla | direção    |
      | W     | cima       |
      | A     | a esquerda |
      | S     | baixo      |
      | D     | a direita  |

  # S+D, não W+D: logo acima do ponto de nascimento fica a fonte. Com W+D o
  # guaxinim bate nela e passa a andar só para a direita, e o cenário passaria
  # mesmo se a diagonal valesse frente/costas (achado na mutação da T8).
  Cenário: Na diagonal vale o lado
    Dado que "rocky" e "bandit" estão na Praça
    Quando "rocky" segura as teclas "S" e "D"
    Então "bandit" vê "rocky" olhando para a direita

  Cenário: Quem acabou de entrar olha para baixo
    Dado que "rocky" e "bandit" estão na Praça
    Então "bandit" vê "rocky" olhando para baixo

  Cenário: Continuar olhando para onde andava ao parar
    Dado que "rocky" está segurando a tecla "D"
    Quando "rocky" solta a tecla "D"
    Então "rocky" está parado olhando para a direita

  Cenário: Animação de andar
    Dado que "rocky" e "bandit" estão na Praça
    Quando "rocky" segura a tecla "S"
    Então "bandit" vê "rocky" andando

  Cenário: Animação parado
    Dado que "rocky" e "bandit" estão na Praça
    Então "bandit" vê "rocky" parado

  Cenário: Anel no chão do próprio guaxinim
    Dado que "rocky" e "bandit" estão na Praça
    Então "rocky" vê o anel embaixo do próprio guaxinim
    E "rocky" não vê anel embaixo do guaxinim de "bandit"
