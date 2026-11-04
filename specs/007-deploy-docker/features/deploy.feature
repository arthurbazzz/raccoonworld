# language: pt
# Roda contra a stack Docker Compose no ar (TestDeployStack, com RW_STACK_URL).
# O cenário "Push na main atualiza a VPS" da spec é verificado pelo workflow deploy.yml,
# que confere a tag do commit e o /health depois do deploy.
@RW-007
Funcionalidade: Jogo em produção com Docker Compose

  Contexto:
    Dado que a stack de produção está no ar

  Cenário: Página do jogo pelo endereço público
    Quando o navegador abre "/"
    Então a Praça carrega

  Cenário: Jogador conecta pelo WebSocket através do proxy
    Quando "rocky" abre o jogo pelo endereço público
    Então "rocky" entra na Praça

  Cenário: Mapa servido pelo backend através do proxy
    Quando alguém pede "/maps/praca"
    Então recebe o mapa da Praça

  Cenário: Só a porta 80 fica exposta
    Então a única porta publicada pela stack é a 80
    E o Postgres não aceita conexão de fora da stack

  Cenário: Dados do banco sobrevivem a reinício
    Dado que existe a tabela "probe" com uma linha no Postgres
    Quando a stack é derrubada e sobe de novo
    Então a tabela "probe" ainda tem a linha

  Cenário: Origem desconhecida é recusada
    Quando um navegador em outra origem tenta abrir o WebSocket
    Então a conexão é recusada
