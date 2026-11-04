# RW-007 — Deploy em produção com Docker Compose

Status: **Aprovada** (2026-11-04) · Criada: 2026-11-04 · Depende de: RW-001

## Problema

O jogo só roda na máquina de quem desenvolve (`go run` + `npm run dev`). Não há como colocá-lo
numa VPS para outras pessoas jogarem, e cada atualização exigiria passos manuais.

## Objetivo

Uma **stack Docker Compose** que sobe o jogo inteiro numa VPS própria, acessível pelo **IP
público**, e um **pipeline no GitHub Actions** que testa, constrói as imagens, publica no GHCR e
atualiza a VPS a cada push na `main`.

## Fora de escopo

- Domínio próprio e HTTPS/TLS (Q2). O acesso é `http://<IP>` e `ws://<IP>/ws`.
- Uso do Postgres pelo servidor: conexão, repositórios e migrations (Q3). O banco só fica pronto
  na stack.
- Backup do banco, monitoramento, alertas e logs centralizados.
- Escalar horizontalmente (mais de uma instância do `server`): as salas vivem em memória.
- Ambiente de homologação separado.

## Usuários

- **Dono do projeto**, que faz o deploy.
- **Visitante**, que joga pelo navegador usando o IP.

## Regras

### Imagens

- RN1. Duas imagens, cada uma com `Dockerfile` multi-stage na sua pasta:
  - **server** (`server/Dockerfile`): binário Go estático numa imagem mínima, rodando como
    usuário não-root;
  - **web** (`web/Dockerfile`): Next.js em modo `standalone`, com os sprites gerados no build
    (`npm run sprites`), rodando como usuário não-root.
- RN2. A URL do WebSocket do cliente (`NEXT_PUBLIC_WS_URL`) é gravada no build da imagem
  **web**, recebida como argumento de build.
- RN3. As imagens são publicadas como **privadas** no GHCR (`ghcr.io/<dono>/raccoonworld-server`
  e `-web`) com duas tags: o SHA do commit e `latest`.

### Stack (`deploy/compose.yaml`)

- RN4. Quatro serviços: **traefik**, **server**, **web** e **postgres**.
- RN5. Só o **Traefik** expõe porta para fora, a **80**. Roteamento:
  - `/ws`, `/maps/*` e `/health` vão para o **server**;
  - todo o resto vai para o **web**.
  As rotas são declaradas por labels nos serviços (Q1). O dashboard do Traefik não fica exposto.
- RN6. O **Postgres** usa um volume nomeado (os dados sobrevivem a `down`/`up`), não expõe
  porta e tem healthcheck. O `server` só sobe depois do banco `healthy` e já recebe
  `DATABASE_URL`, mesmo sem usá-la ainda.
- RN7. Todos os serviços têm `restart: unless-stopped`. `server` e `web` têm healthcheck.
- RN8. O `server` tem tempo de parada suficiente para o desligamento gracioso que já existe.
- RN9. Variáveis e segredos ficam num `deploy/.env` **fora do git**. O repositório versiona só
  `deploy/.env.example`. Em produção, `RW_CUPS_FAST`, `RW_RUMMY_FAST`, `RW_RUMMY_SEED` e
  `RW_DEBUG` ficam desligadas; `RW_ALLOWED_ORIGINS` contém o IP público.
- RN10. A mesma stack sobe localmente com as imagens construídas na máquina
  (`docker compose up --build`), para testar antes do deploy.

### Pipeline (GitHub Actions)

- RN11. **Pull request e push na `main`:** roda lint e testes do backend e do frontend. Se algo
  falhar, nada é publicado.
- RN12. **Push na `main` com testes verdes:** constrói e publica as duas imagens (RN3) e então
  faz o deploy na VPS por SSH (`docker compose pull` + `docker compose up -d`).
- RN13. IP da VPS, usuário, chave SSH e o conteúdo do `.env` de produção ficam em
  **GitHub Secrets**. Nada sensível aparece nos logs do workflow.
- RN14. Um deploy que falha não derruba a versão que está no ar. O workflow confere o
  `/health` depois do `up` e termina com erro se ele não responder.

### Preparação da VPS

- RN15. Um guia em `deploy/README.md` descreve o preparo único da VPS: instalar Docker, criar
  usuário de deploy com chave SSH, liberar a porta 80 no firewall e fazer login no GHCR com um
  token de leitura de pacotes.

## Critérios de aceitação

Rascunho dos cenários. A versão final vai para `features/` na fase de tasks. Eles rodam contra a
stack local (RN10), exceto o último, que é verificado no workflow.

```gherkin
# language: pt
Funcionalidade: Jogo em produção com Docker Compose

  Contexto:
    Dado que a stack de produção está no ar

  Cenário: Página do jogo pelo IP
    Quando o navegador abre "http://<IP>/"
    Então a Praça carrega

  Cenário: Jogador conecta pelo WebSocket através do proxy
    Quando "rocky" abre o jogo pelo IP
    Então "rocky" entra na Praça
    E a conexão usa "ws://<IP>/ws"

  Cenário: Mapa servido pelo backend através do proxy
    Quando alguém pede "http://<IP>/maps/praca"
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

  Cenário: Push na main atualiza a VPS
    Dado que os testes passaram no workflow
    Quando o workflow publica as imagens e faz o deploy
    Então a VPS roda as imagens com a tag do commit
    E "/health" responde com sucesso
```

## Decisões

- **Q1. Proxy: Traefik** (o dono já conhece), com rotas por labels no compose.
- **Q2. Sem domínio e sem TLS por enquanto:** acesso pelo IP público, só para teste. Trocar
  para domínio + Let's Encrypt depois exige só mudar as rotas do Traefik e as variáveis.
- **Q3. Postgres só na stack:** sem conexão, repositórios ou migrations nesta spec.
- **Q4. CI/CD no GitHub Actions + GHCR**, repositório **privado** (cota do plano Free:
  ~2.000 min/mês de Actions e 500 MB de pacotes).
- **Q5. CI sem e2e por enquanto:** o pipeline roda lint, `go test ./...` (inclui godog) e
  Vitest. O `test:e2e` (playwright-bdd) fica fora para poupar tempo e cota.
- **Q6. Os critérios de aceitação viram um script de smoke** automático, que sobe a stack local
  e confere os cenários. Ele roda localmente e no CI, antes de publicar as imagens.
