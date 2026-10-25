# Raccoons World — Constitution

Versão: 1.3.0 · Ratificada: 2026-10-25 · Última emenda: 2026-10-25 (idioma do código)

Esta constituição define as regras inegociáveis do projeto. Ela prevalece sobre
qualquer outra instrução, spec ou conveniência de implementação.

## 1. Propósito

Raccoons World é um jogo virtual focado em interação entre usuários, com um
ambiente interativo que hospeda minigames.
As referências de game design ficam em [`references.md`](references.md).

## 2. Princípios

### I. Spec antes de código (inegociável)
Toda alteração no projeto — feature, correção, refatoração ou mudança de
infraestrutura — começa com uma spec em `specs/<NNN>-<nome-curto>/`.
Código sem spec correspondente não é aceito.

### II. Fluxo obrigatório
`spec.md` → `plan.md` → `tasks.md` → implementação → verificação.
- **spec.md**: o quê e por quê (problema, usuários, critérios de aceitação). Sem detalhes técnicos.
- **plan.md**: como (arquitetura, contratos de API, modelo de dados, decisões e alternativas descartadas).
- **tasks.md**: passos pequenos, ordenados e verificáveis, cada um com ID (`T1`, `T2`…).

Cada etapa só começa depois que a anterior foi revisada e aprovada pelo dono do projeto.

### III. Clean Architecture
Dependências apontam sempre para dentro:
`domain` ← `usecase` ← `adapter` (http, ws, repository) ← `infra` (framework, banco, config).
- O domínio não importa framework, banco nem bibliotecas de transporte.
- Regras de negócio vivem em casos de uso, nunca em handlers, controllers ou componentes de UI.
- A comunicação entre camadas é feita por interfaces definidas na camada interna.

### IV. Testes primeiro, em BDD (inegociável)
Os testes seguem Behavior-Driven Development:
- Os critérios de aceitação da `spec.md` são escritos como cenários Gherkin
  (`Funcionalidade` / `Cenário` / `Dado` / `Quando` / `Então`), em português (`# language: pt`).
- Cada cenário vira um arquivo `.feature` executável **antes** da implementação. O cenário
  precisa falhar primeiro e passar depois.
- Backend: os `.feature` são executados com **godog** (Cucumber para Go) e os steps chamam os
  casos de uso ou a API, nunca detalhes internos.
- Frontend: os fluxos de usuário em `.feature` são executados com **playwright-bdd**.
- Testes unitários (`go test`, sem banco nem rede) complementam o BDD no domínio
  e nos casos de uso. Os repositórios têm testes de integração contra Postgres real (container).
- Uma task só é concluída quando todos os cenários da spec passam.

### V. Servidor autoritativo
Num jogo multiusuário, o backend é a fonte da verdade sobre o estado do mundo,
os minigames e as pontuações. O cliente envia intenções e o servidor valida e decide.
Nenhuma regra de jogo depende só do frontend.

### VI. Contratos explícitos
As APIs (HTTP e tempo real) são definidas no `plan.md` antes de serem implementadas:
OpenAPI para REST e mensagens versionadas e tipadas para WebSocket.
Quebrar um contrato exige uma spec própria.

### VII. Simplicidade
Implementar só o que a spec pede. Nada de abstrações especulativas.
Uma dependência nova só entra com justificativa registrada no `plan.md`.

### VIII. Idioma
- **Código em inglês**: identificadores (variáveis, funções, tipos, pacotes, arquivos),
  comentários, mensagens de log e de erro internas, chaves de JSON e nomes de tabelas e colunas.
- **Documentação de produto em português**: specs, plans, tasks, a constituição e os
  cenários Gherkin (`.feature`, `# language: pt`), porque descrevem o negócio.
- Os steps BDD seguem o texto em português do `.feature`, mas a implementação
  (funções e variáveis) é em inglês.
- Textos exibidos ao jogador são conteúdo de UI e ficam fora desta regra. Por enquanto,
  em português.

### IX. Rastreabilidade
Branches, commits e PRs referenciam a spec e a task (ex.: `RW-003/T2`).
Se a implementação divergir da spec, a spec é atualizada no mesmo PR.

## 3. Stack

| Camada    | Tecnologia                         |
|-----------|------------------------------------|
| Frontend  | Next.js (TypeScript)               |
| Backend   | Go + Gin                           |
| Testes    | BDD: godog (Go), playwright-bdd (Next.js) |
| Banco     | PostgreSQL                         |

Restrições:
- As migrations do banco são versionadas no repositório. Nada de alterar o schema manualmente.
- Os segredos ficam em variáveis de ambiente e nunca são commitados.
- O frontend só acessa o banco através do backend Go.
- O Gin fica restrito às camadas `adapter`/`infra`. Os tipos do Gin (`*gin.Context` etc.)
  nunca chegam aos casos de uso nem ao domínio.

## 4. Definição de Pronto

- [ ] Critérios de aceitação da spec atendidos
- [ ] Cenários BDD (`.feature`) da spec passando
- [ ] Testes passando, lint e formatação limpos (`gofmt`/`golangci-lint`, `eslint`)
- [ ] Nenhuma violação de camada da Clean Architecture
- [ ] spec/plan/tasks atualizados com o que foi realmente feito

## 5. Governança

- Emendas exigem uma proposta escrita com o motivo e um incremento de versão:
  MAJOR quando um princípio é removido ou alterado, MINOR quando um é adicionado, PATCH para esclarecimentos.
- Uma spec que conflite com esta constituição deve ser corrigida, não contornada.
- Na dúvida entre seguir a regra e "ir mais rápido", segue-se a regra e se registra a questão.

## 6. Decisões em aberto

- [x] Framework HTTP do backend Go → **Gin** (v1.1.0)
- [x] Abordagem de testes → **BDD com Gherkin** (v1.1.0)
- [x] Tempo real → **WebSocket** com mensagens JSON versionadas (`type`, `v`) e
      **servidor autoritativo por sala**. O cliente envia intenções e o servidor valida
      e envia o resultado para a sala. Pub/sub entre instâncias só quando houver escala (v1.2.0)
- [ ] Autenticação: **adiada**. Direção provável: Firebase Auth com Sign in with Apple
      e cadastro manual (e-mail/senha). Até lá, o código depende só de uma interface
      de identidade no domínio, sem acoplar a nenhum provedor.
- [ ] Monorepo e estrutura de pastas
