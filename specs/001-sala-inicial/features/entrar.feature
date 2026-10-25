# language: pt
@RW-001
Funcionalidade: Entrar na Praça
  Como visitante
  Quero entrar na Praça com um apelido
  Para encontrar outros jogadores

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

  Esquema do Cenário: Apelido em uso com outra capitalização
    Dado que "rocky" está na Praça
    Quando outro visitante tenta entrar com o apelido "<apelido>"
    Então a entrada é recusada com o motivo "apelido em uso"

    Exemplos:
      | apelido |
      | Rocky   |
      | ROCKY   |

  Esquema do Cenário: Apelido inválido
    Quando o visitante tenta entrar com o apelido "<apelido>"
    Então a entrada é recusada com o motivo "apelido inválido"

    Exemplos:
      | apelido               |
      | a                     |
      | ab                    |
      | guaxinim_muito_grande |
      | rocky!                |
      | com espaço            |

  Cenário: Sala cheia
    Dado que a Praça está com a capacidade máxima
    Quando o visitante tenta entrar com o apelido "bandit"
    Então a entrada é recusada com o motivo "sala cheia"

  Cenário: Outros jogadores veem a chegada
    Dado que "rocky" está na Praça
    Quando "bandit" entra na Praça
    Então "rocky" é avisado de que "bandit" entrou
