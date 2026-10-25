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
7. Gin só nos handlers e adapters. Os casos de uso recebem DTOs próprios, nunca `*gin.Context`.
8. Decisões listadas como "em aberto" na constituição (ex.: framework Go) são
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
- Frontend: `npm run dev` · `npm test` · `npx bddgen && npx playwright test` (BDD) · `npm run lint`
- Backend: `go run ./cmd/...` · `go test ./...` (inclui godog) · `golangci-lint run`
