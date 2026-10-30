# CLAUDE.md — Raccoons World

Leia e siga `specs/constituion.md`. Ela prevalece sobre este arquivo.

## Regras de trabalho

1. **Não há spec aprovada?** Não escreva código. Ajude a escrever a spec primeiro.
2. Siga o fluxo `spec.md` → `plan.md` → `tasks.md` e espere a aprovação entre as etapas.
3. Implemente uma task de `tasks.md` por vez e marque-a ao concluir.
4. Se a spec estiver ambígua, pergunte em vez de presumir, e registre a decisão na spec.
5. Se o escopo mudar durante a implementação, pare e proponha a atualização da spec.
6. **BDD**: escreva os critérios de aceitação da spec como cenários Gherkin em pt,
   crie os `.feature` e os steps **antes** do código e confirme que falham primeiro.
   Backend com godog, frontend com playwright-bdd.
7. **Código em inglês** (identificadores, comentários, logs, JSON). Specs e `.feature`
   ficam em português.
8. Gin só nos handlers e adapters. Os casos de uso recebem DTOs próprios, nunca `*gin.Context`.
9. Decisões listadas como "em aberto" na constituição (ex.: framework Go) são
   discutidas com o usuário antes de qualquer escolha.

## Estrutura de specs

```
specs/
  constituion.md
  001-<nome-curto>/
    spec.md
    plan.md
    tasks.md
    features/        # cenários Gherkin (.feature) desta spec
```

Exemplo de cenário:

```gherkin
# language: pt
Funcionalidade: Entrar em um minigame
  Cenário: Jogador entra em uma sala com vaga
    Dado que o jogador "rocky" está autenticado
    E existe a sala "corrida-01" com 3 de 4 vagas ocupadas
    Quando "rocky" pede para entrar na sala "corrida-01"
    Então "rocky" passa a fazer parte da sala
    E a sala fica com 4 de 4 vagas ocupadas
```

## Comandos

<!-- Preencher quando o projeto tiver código -->
- Frontend (em `web/`): `npm run dev` · `npm test` (Vitest) · `npm run test:e2e` (playwright-bdd,
  cenários `@ui`/`@e2e`; sobe os dois servidores sozinho) · `npm run lint` · `npm run format:check`
- Sprites (RW-004): a arte fonte fica em `web/art/`. `npm run sprites` gera `web/public/sprites/`
  (fora do git) e roda sozinho antes de `dev`, `build` e `test:e2e`; se a imagem sair do formato
  3×6, esses comandos param. Os scripts em `web/scripts/` rodam com o TypeScript nativo do Node:
  sem `enum` e sem parâmetros de construtor com modificador; importe com a extensão `.ts`.
- O BDD de UI usa um único servidor compartilhado. Feche abas do jogo abertas em outro navegador
  antes de rodar: elas reconectam sozinhas e ocupam os apelidos dos cenários.
- Backend (em `server/`): `go run ./cmd/server` · `go test ./...` (inclui godog) ·
  `go tool -modfile=tools/go.mod golangci-lint run ./...` · `go tool -modfile=tools/go.mod golangci-lint fmt ./...`
- BDD do backend: `go test ./features/...`. Ele roda em modo strict, então steps pendentes falham:
  uma spec nova começa vermelha e fica verde conforme as tasks avançam. São duas suítes:
  `TestFeatures`, contra os casos de uso com relógio falso, e `TestSmokeOverWebSocket`, que roda os
  cenários `@smoke` contra o servidor real (`internal/app`). Para filtrar
  cenários, use `GODOG_TAGS` (ex.: `$env:GODOG_TAGS="@RW-001"`).
- Três copos (RW-005): o domínio fica em `server/internal/domain/cups` (máquina de estados pura,
  avançada pelo tick da sala) e o cliente em `web/src/game/cups/` (`timeline.ts` diz onde está
  cada copo em qualquer instante; painel e barraca usam a mesma). O `test:e2e` sobe o servidor
  com `RW_CUPS_FAST=1`. Nos steps de navegador, ande em toques curtos (segurar a tecla passa do
  ponto) e espere fases rápidas com `waitForFunction` + `raf`, não com `expect`.
- O golangci-lint também barra imports que violam a Clean Architecture (depguard em `server/.golangci.yml`).
  Não desative essas regras: corrija a dependência.
