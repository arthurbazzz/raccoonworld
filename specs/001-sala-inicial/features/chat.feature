# language: pt
@RW-001
Funcionalidade: Chat na Praça
  Como jogador
  Quero enviar mensagens curtas
  Para conversar com quem está na Praça

  Contexto:
    Dado que a sala "Praça" está aberta

  Cenário: Mensagem de chat
    Dado que "rocky" e "bandit" estão na Praça
    Quando "rocky" envia a mensagem "olá!"
    Então "bandit" vê o balão "olá!" sobre "rocky"

  Cenário: Mensagens rápidas demais
    Dado que "rocky" e "bandit" estão na Praça
    E "rocky" acabou de enviar uma mensagem
    Quando "rocky" envia outra mensagem em menos de 1 segundo
    Então a segunda mensagem é descartada

  Cenário: Mensagem longa demais
    Dado que "rocky" e "bandit" estão na Praça
    Quando "rocky" envia uma mensagem com 121 caracteres
    Então a mensagem é descartada
