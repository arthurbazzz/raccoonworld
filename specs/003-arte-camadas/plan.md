# RW-003 — Plano técnico: Arte do guaxinim em camadas

Status: **Rascunho**, aguardando aprovação · Spec: [spec.md](spec.md) · Criado: 2026-10-25

## 1. Visão geral

Só o frontend muda, e só o **desenho** do guaxinim. A direção, a animação, o anel e a lista
acessível da RW-002 ficam como estão (RA1). A cena continua chamando `RaccoonView` com a
mesma interface: `setFacing`, `animate` e `ring`.

O arquivo único `raccoonArt.ts` vira uma pasta, com uma responsabilidade por arquivo:

```
web/src/game/raccoon/
  palette.ts      paleta base + tons de sombra e contorno derivados (RA5, RA7)
  shapes.ts       formas com contorno e sombra (elipse, círculo, orelha, rabo…)
  rig.ts          as 3 poses: camadas (RA2) e encaixes (RA3) de cada uma
  accessories.ts  tipo Accessory + o chapéu de exemplo (RA4)
  RaccoonView.ts  o Container que a cena usa (setFacing, animate, equip)
  index.ts        reexporta a API pública
```

`raccoonArt.ts` passa a só reexportar `raccoon/`, e some quando ninguém mais importar dele.

## 2. Paleta e sombreado (RA5, RA7)

```ts
interface RaccoonPalette {
  fur: number; furLight: number; mask: number; earInner: number;
  tailRing: number; nose: number; eye: number;
}
```

- Os tons derivados (`shade(cor) = cor × 0,75`, contorno = `shade(fur) × 0,45`) são **calculados**,
  não guardados. Trocar só o `fur` troca o pelo, a sombra e o contorno juntos.
- **Técnica da sombra**, sem degradê:
  1. desenha a forma inteira no tom escuro, com o contorno;
  2. desenha por cima a mesma forma em cor base, um pouco menor e deslocada para cima e para a esquerda.

  Sobra uma "meia-lua" escura embaixo e à direita, que é a sombra simples da referência.
  É barato, porque são 2 formas por parte, e fica nítido em qualquer escala.

## 3. Camadas e poses (RA2)

Cada pose é um `Container` (o `body`, que espelha) com um `Container` por camada, nesta ordem:

| Pose | Ordem (de trás para a frente) |
|---|---|
| Frente | rabo · pernas · corpo · braços · cabeça · rosto |
| Costas | pernas · corpo · braços · cabeça · rabo |
| Lado | rabo · perna de trás · corpo · perna da frente · braço · cabeça · rosto |

**Proporções da referência** (para cerca de 45 px de altura):
- **Cabeça:** raio de ~13 px, centro a −27. Ocupa ~55% da altura.
- **Corpo:** elipse de ~9×9 px, centro a −10, com a barriga creme.
- **Braços e pernas:** curtos e arredondados. Os braços balançam ao andar, em oposição às pernas.
- **Rosto de frente:** a máscara é uma "gota" dupla em volta dos olhos, com a faixa creme acima.
  Os olhos são grandes, com brilho branco, e há nariz e boquinha. O rosto de lado é um perfil com focinho.
- **Costas (Q2):** nuca com uma mancha mais clara, orelhas com contorno e interior visível, e o
  rabo com volume (sombra + anéis) por cima do corpo.

A animação da RW-002 continua funcionando sobre as mesmas partes: `body`, `torso` (agora o corpo),
`head`, `tail` e `legs`. Os braços entram como `arms`, com o par oposto ao das pernas.

## 4. Encaixes e acessórios (RA3, RA4)

```ts
type Slot = "chapeu" | "oculos" | "camiseta" | "calcado";

interface SlotAnchor {
  part: "head" | "torso" | "legs";  // a camada onde o acessório é preso
  x: number; y: number;             // posição relativa a essa camada
  above: boolean;                   // por cima (true) ou por baixo da camada
}

interface Accessory {
  id: string;
  slot: Slot;
  draw: Record<Pose, (g: Graphics, palette: RaccoonPalette) => void>;
}
```

- Cada pose define um `SlotAnchor` por encaixe no `rig.ts`.
- `RaccoonView.equip(acessório)` desenha o acessório e o põe **como filho da camada do encaixe**.
  Assim ele herda de graça o balanço, a respiração e o espelhamento. `unequip(slot)` remove.
- Quando a pose muda, a view **reequipa** tudo o que estava equipado na pose nova.
- **Chapéu de exemplo:** um chapéu-coco, ou de festa, desenhado em 3 poses, no encaixe `chapeu`.
  Ele só é usado na galeria (Q4).
- Uma roupa nova é só um objeto `Accessory` novo, sem mudar o `RaccoonView` (RA4).

## 5. Galeria de desenvolvimento (RA8)

- `web/src/app/galeria/page.tsx`: um client component que mostra uma grade com as poses
  (frente, costas, direita e esquerda), andando e parado, sem e com o chapéu, a 1× e a 4×.
- **Produção:** a página chama `notFound()` quando `process.env.NODE_ENV === "production"`, então
  o build responde **404**. Isso é conferido no fechamento com `next build && next start`.

## 6. Testes

| O quê | Ferramenta | Como |
|---|---|---|
| Ordem das camadas (RA2) | Vitest | Índice de cada camada no `body`, por pose |
| Encaixes (RA3) | Vitest | O chapéu é filho da camada `head`, sobe com o balanço do andar, espelha à esquerda e é reequipado ao trocar de pose |
| Acessório como dados (RA4) | Vitest | Um acessório de teste, criado só no teste, encaixa sem mudar o `RaccoonView` |
| Paleta e sombra (RA5, RA7) | Vitest | Com outra cor de `fur`, as instruções de preenchimento do `GraphicsContext` usam a cor base e o tom derivado |
| Apoio nos pés e tamanho (RA1, RA6) | Vitest | Limites da arte: pés em y≈0, altura entre 40 e 48 px, centrado |
| Comportamento da RW-002 (RA1) | Vitest + playwright-bdd | Os testes de `raccoonArt` que testam comportamento continuam; os 14 cenários de navegador sem mudança |
| Qualidade visual (RA6) | **Dono do projeto** | Aprovação na `/galeria` |

Alguns testes da RW-002 que verificavam **detalhes do desenho antigo** vão mudar (ex.: "2 patas
de frente", "rabo atrás do torso"). Os que verificam **comportamento** (pose por direção,
espelhar só o corpo, apoio nos pés, animação) continuam valendo.

## 7. Dependências novas

Nenhuma.

## 8. Riscos

- **Desempenho:** cada parte passa a ter 2 formas (sombra + base) mais o contorno. Com 10
  guaxinins isso continua barato, e as poses só são redesenhadas quando a direção muda.
- **Legibilidade a 1×** (RA6): detalhes como o brilho dos olhos e a boquinha podem sumir em 45 px.
  A galeria mostra o tamanho real justamente para decidir isso.
- **Semelhança com a referência:** o sombreado em código é mais "chapado" do que o da imagem.
  Já foi combinado na spec (Q1).

## 9. Decisões deste plano que precisam do seu OK

1. **Pasta `game/raccoon/`** no lugar do arquivo único `raccoonArt.ts`.
2. **Sombra pela técnica das 2 formas** (escura embaixo, base deslocada por cima), em vez de degradê.
3. **Acessório como filho da camada do encaixe**, para herdar animação e espelhamento.
4. **Galeria com `notFound()` em produção** (a página existe no código, mas responde 404 no build).
