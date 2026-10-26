# RW-002 — Visual do guaxinim

Status: **Aprovada** (2026-10-25) · Criada: 2026-10-25 · Depende de: RW-001

## Problema

Na RW-001 o guaxinim é um placeholder feito com formas simples (círculos e retângulos),
sempre de frente e parado, mesmo quando anda. Ele não parece estar vivo:
- não vira para o lado para onde anda;
- não tem animação de andar nem de respirar;
- é difícil reconhecer o seu guaxinim no meio dos outros.

Num jogo social, o personagem é a "cara" do jogador. As referências (Club Penguin, Habbo,
Mundo do Gaturro, Party Parrot World) têm personagens expressivos e reconhecíveis.

## Objetivo

O guaxinim passa a ter um visual de personagem:
- arte própria na perspectiva 3/4 (RN6 da RW-001), **desenhada em código** (vetorial, com
  PixiJS Graphics), com a máscara e o rabo listrado, sem arquivos de imagem;
- **olha para a direção do movimento**, em **4 lados**: frente, costas, esquerda e direita;
- **anima ao andar** e tem uma animação sutil **parado** (respiração ou piscar);
- o **seu** guaxinim é fácil de achar no meio dos outros.

## Fora de escopo

- Customização (cores, roupas, acessórios). Fica para uma spec própria, que vai precisar
  persistir dados e, por isso, depende da autenticação.
- Emotes, gestos e animações disparadas pelo jogador.
- Arte do cenário (chão, árvores, fonte).
- Mudanças no servidor ou no protocolo. A direção e a animação são deduzidas no cliente a
  partir das posições que o servidor já envia.

## Usuários

- **Jogador**: vê o próprio guaxinim e os dos outros na Praça.

## Regras

- RV1. A **direção** do guaxinim é a do último movimento, em 4 lados:
  - andando para baixo, aparece de **frente**; para cima, de **costas**;
  - para a esquerda ou para a direita, aparece de **lado**;
  - na diagonal vale o **lado** (esquerda ou direita), porque é o que mais aparece na tela;
  - ao parar, continua olhando para onde andava. Quem acabou de entrar aparece de frente.
- RV2. Enquanto a posição muda, toca a animação de **andar**. Quando para, depois de no máximo
  150 ms toca a animação **parado** (respiração sutil). Não há outras animações nesta spec.
- RV3. O guaxinim de quem está jogando tem um **anel no chão**, embaixo dos pés, que os outros
  não têm, além do nome em amarelo que já existe.
- RV4. O visual não muda a colisão: a hitbox de 20×12 px nos pés continua igual, e o ponto
  de apoio do sprite continua nos pés (ordenação 3/4).
- RV5. Todos os jogadores veem a mesma direção e o mesmo estado (andando/parado) para cada
  guaxinim, porque tudo é deduzido das posições do servidor.

## Critérios de aceitação

```gherkin
# language: pt
Funcionalidade: Visual do guaxinim

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

  Cenário: Na diagonal vale o lado
    Dado que "rocky" e "bandit" estão na Praça
    Quando "rocky" segura as teclas "W" e "D"
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
```

Como a arte é desenhada no canvas, o teste não "olha" a imagem. Os cenários conferem o
estado visual (direção, andando/parado, destaque), que o jogo também expõe na lista
acessível de jogadores, como a RW-001 já faz com a posição.

## Decisões

- **Q1. Arte:** vetorial, **desenhada em código** com PixiJS Graphics. Sem arquivos de imagem.
- **Q2. Direções:** **4 lados** (frente, costas, esquerda e direita). Na diagonal vale o lado (RV1).
- **Q3. Animações:** só **andar** e **parado** (RV2).
- **Q4. Destaque:** **anel no chão** embaixo do próprio guaxinim (RV3).
