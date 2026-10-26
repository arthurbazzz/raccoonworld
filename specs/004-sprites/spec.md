# RW-004 — Guaxinim em sprites (imagem)

Status: **Rascunho**, aguardando aprovação · Criada: 2026-10-25 · Substitui: RW-003 · Depende de: RW-002

## Problema

A RW-003 tentou chegar ao estilo da imagem de referência desenhando o guaxinim **em código**.
O dono do projeto avaliou o resultado e **não gostou do formato**. A arte em código tem um teto
de qualidade que não chega à ilustração.

Em vez disso, o dono gerou com IA uma **folha de sprites** (sprite sheet) do guaxinim, no estilo
desejado: 3 direções (frente, costas e lado), parado e andando, com fundo transparente.

## Objetivo

O guaxinim do jogo passa a ser **desenhado a partir dessa imagem**, mantendo tudo o que a RW-002
já faz:
- olha para 4 lados (a esquerda é o lado espelhado);
- anima ao andar e respira parado;
- tem o anel no chão do próprio jogador;
- fica apoiado nos pés, com a mesma colisão e a mesma ordenação 3/4.

## Fora de escopo

- **Acessórios e roupas** (chapéu, etc.). Ficam para uma spec futura. A prévia com o chapéu mostrou
  que é viável posicionar uma folha de roupa pelo código.
- Customização de cores, que com imagens exige outra abordagem.
- Gerar ou corrigir a arte. A imagem é usada como o dono entregar.
- Mudanças no servidor, no protocolo ou no cenário.

## A imagem

- Arquivo fornecido pelo dono: PNG de **768×384 px** com fundo transparente.
- **3 fileiras:** frente (olhando para baixo), costas (olhando para cima) e lado (olhando para a direita).
- **6 colunas:** colunas 1–2 **parado** e colunas 3–6 **andando** (ciclo de 4 quadros).
- Medido em 2026-10-25: as **fileiras não seguem a grade de 128 px** (a de frente ocupa
  y=22–136, por exemplo). Dentro de cada fileira, os pés variam no máximo ±2 px. Cada guaxinim
  tem 98–112 px de altura.

## Regras

- RS1. **Mesmo comportamento da RW-002:** a direção (4 lados, diagonal pelo lado), "andando"
  nos últimos 150 ms, o anel, o apoio nos pés e a hitbox de 20×12 px. Nada disso muda.
- RS2. **Quadros detectados automaticamente:** cada quadro é encontrado pelas faixas
  transparentes entre eles, e não por uma grade fixa. Se a imagem não tiver exatamente
  3 fileiras × 6 colunas, o processo **falha com uma mensagem clara**, e nunca desenha um
  guaxinim cortado.
- RS3. **Alinhamento pelos pés:** todo quadro é posicionado pelo centro da base (os pés), então
  as pequenas variações da IA não fazem o guaxinim "pular" de um quadro para outro.
- RS4. **Tamanho no jogo:** o guaxinim fica com cerca de **46 px de altura** na sala, a mesma
  altura da RW-003. O nome e o balão continuam acima da cabeça.
- RS5. **Animação:**
  - **andando:** os 4 quadros de andar em sequência, no ritmo do passo;
  - **parado:** os 2 quadros de parado alternando devagar, como uma respiração.
- RS6. **Trocar a arte é trocar o arquivo:** uma imagem nova com o mesmo layout (3×6, fundo
  transparente) funciona sem mudar código.
- RS7. A **galeria** `/galeria` mostra o guaxinim em sprites: 4 direções, parado e andando.

## Critérios de aceitação

- Os **14 cenários de navegador** (RW-001 e RW-002) continuam passando, sem mudança (RS1).
- Testes unitários com imagens de teste pequenas cobrem:
  - a detecção dos quadros, inclusive fileiras fora da grade;
  - o erro para uma imagem com quadros faltando;
  - o alinhamento pelos pés;
  - a escolha do quadro por direção, estado e tempo.
- A **qualidade visual** (nitidez e animação) é aprovada pelo dono na galeria.

## Questões em aberto

- **Q1. Onde recortar os quadros:**
  - (a) **na hora de gerar o site** (recomendado): um script lê a imagem, acha os quadros e grava
    uma folha alinhada mais um JSON com as posições. O jogo só carrega o resultado, e um erro
    na imagem aparece antes de ir para o ar;
  - (b) **no navegador**, a cada carregamento: mais simples de trocar a imagem, mas mais lento
    para abrir e com o erro aparecendo só para o jogador.
- **Q2. Arte em código da RW-003:** remover (a pose de frente, a paleta e as formas) ou manter
  como reserva caso a imagem não carregue? Recomendo **remover** e, se a imagem falhar, mostrar
  a mensagem de erro que o jogo já tem.
- **Q3. Ritmo do andar:** os 4 quadros acompanham o passo atual (6 passos/s, ou seja ~12 quadros/s)?
  Ou prefere mais lento?
