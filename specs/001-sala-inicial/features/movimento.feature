# language: pt
@RW-001
Funcionalidade: Andar pela Praça com WASD
  Como jogador
  Quero mover meu guaxinim com o teclado
  Para explorar a Praça junto com os outros

  Contexto:
    Dado que a sala "Praça" está aberta

  @e2e
  Cenário: Andar com WASD
    Dado que "rocky" e "bandit" estão na Praça
    Quando "rocky" segura a tecla "D"
    Então "rocky" se move para a direita
    E "bandit" vê "rocky" se mover para a direita

  Esquema do Cenário: Cada tecla move em uma direção
    Dado que "rocky" está na Praça
    Quando "rocky" segura a tecla "<tecla>"
    Então "rocky" se move para <direção>

    Exemplos:
      | tecla | direção    |
      | W     | cima       |
      | A     | a esquerda |
      | S     | baixo      |
      | D     | a direita  |

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

  @ui
  Cenário: Digitar no chat não move o guaxinim
    Dado que "rocky" está com o campo de chat em foco
    Quando "rocky" digita "wasd"
    Então a posição de "rocky" não muda
