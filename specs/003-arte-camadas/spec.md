# RW-003 — Arte do guaxinim em camadas

Status: **Substituída pela [RW-004](../004-sprites/spec.md)** (2026-10-25) · Criada: 2026-10-25 · Depende de: RW-002

> **Por que foi substituída:** depois da pose de frente (T5), o dono do projeto avaliou que o
> formato do guaxinim desenhado em código não ficou bom. A decisão foi usar a imagem gerada
> por IA (sprites). T1–T5 ficaram commitadas; a T6 (pose de lado) foi descartada antes do commit.
> A RW-004 decide o que destas tasks é reaproveitado.

## Problema

O guaxinim da RW-002 funciona (vira, anda, respira), mas o desenho ainda é simples.
O dono do projeto gerou uma imagem de referência (sprite sheet com frente, costas e lado),
e o objetivo é chegar perto daquele estilo **mantendo a arte em código**.

Além disso, o jogo vai ter **roupas e acessórios** no futuro. Hoje o desenho é um bloco só
por pose, e não há onde "vestir" nada.

## Objetivo

1. **Redesenhar o guaxinim** no estilo da referência, sempre em código (PixiJS Graphics):
   - cabeça grande (cerca de metade da altura), corpo pequeno, braços e pernas curtos;
   - contorno escuro em todas as partes;
   - rosto com a máscara preta grande, a faixa creme acima dos olhos, olhos grandes com
     brilho, nariz pequeno e boquinha;
   - orelhas pontudas com o interior rosado e um tufo de pelo no topo da cabeça;
   - barriga creme e rabo grosso com anéis escuros.
2. **Organizar o desenho em camadas com pontos de encaixe**, para que uma roupa futura seja
   "um desenho por pose, preso ao encaixe certo", sem mexer no corpo.

## Fora de escopo

- Escolher, comprar ou guardar roupas por jogador. Isso depende da autenticação e terá spec própria.
- Roupas no jogo para outros jogadores verem. Nesta spec o protocolo não muda.
- Sprites em imagem (PNG). A referência serve só de guia visual.
- Mudanças no servidor, na direção, na animação (andar e parado) ou no anel. A RW-002 continua valendo.

## Regras

- RA1. **Mesmo comportamento da RW-002:** 4 direções, a esquerda espelhada, andar e parado,
  anel no próprio guaxinim, ponto de apoio nos pés e hitbox de 20×12 px.
- RA2. **Camadas**, de trás para a frente: rabo, pernas, corpo, braços, cabeça, rosto. Na pose de
  costas, o rabo vai na frente do corpo e o rosto não aparece.
- RA3. **Encaixes (slots)**: cada pose define onde ficam `chapeu`, `oculos`, `camiseta` e `calcado`
  (posição e camada). Um acessório preso a um encaixe:
  - sobe e desce com o corpo ao andar;
  - espelha junto quando o guaxinim olha para a esquerda;
  - fica na camada certa (o chapéu na frente da cabeça; de costas, o chapéu continua por cima).
- RA4. **Um acessório é só dados + desenho por pose.** Adicionar uma roupa nova não exige mudar
  o código do guaxinim.
- RA5. **Paleta única** (como na RW-002). Pelo, máscara e contorno continuam trocáveis num só lugar.
- RA6. **Legível a 1×**: no tamanho do jogo (cerca de 45 px de altura) o guaxinim ainda é
  reconhecível, com a máscara e o rabo visíveis.

## Critérios de aceitação

Os 10 cenários de `specs/002-visual-guaxinim/features/visual.feature` e os 4 da RW-001 no navegador
**continuam passando**, sem mudança (RA1).

As regras de camadas e encaixes (RA2 a RA4) são verificadas com testes unitários do desenho,
por exemplo:
- de frente, a ordem das camadas é rabo < pernas < corpo < braços < cabeça < rosto;
- um chapéu de exemplo preso ao encaixe `chapeu` acompanha o balanço da cabeça ao andar e
  espelha na pose da esquerda;
- trocar a paleta muda a cor do pelo em todas as poses.

A **qualidade visual** (RA6 e a semelhança com a referência) é aprovada pelo dono do projeto,
olhando uma galeria com as 3 poses, andando e parado, com e sem o chapéu de exemplo.

## Decisões

- **Q1. Sombreado:** cores chapadas com contorno **e um tom mais escuro em cada parte**, como
  uma sombra simples (RA7).
- **Q2. Pose de costas:** **mais detalhe** que a referência: nuca, orelhas com contorno e rabo
  com volume, para não virar só uma silhueta.
- **Q3. Encaixes:** `chapeu`, `oculos`, `camiseta` e `calcado` bastam para começar.
- **Q4. Acessório de exemplo:** **um chapéu**, que aparece **só na galeria**. No jogo, ninguém usa
  roupa nesta spec (veja "Fora de escopo").
- **Q5. Galeria:** página `/galeria` **só em desenvolvimento**, fora do build de produção (RA8).

## Regras adicionadas pelas decisões

- RA7. **Sombreado:** cada parte tem a cor base e um tom mais escuro no lado oposto à luz (luz
  vinda de cima e da esquerda), além do contorno. Os tons escuros saem da paleta, então trocar
  a cor do pelo troca a sombra junto.
- RA8. **Galeria de desenvolvimento** em `/galeria`, com as 3 poses (a esquerda também), andando e
  parado, com e sem o chapéu, em tamanho real e ampliado. No build de produção a rota não existe
  (responde 404).
