package features

import "github.com/cucumber/godog"

// direction matches the Portuguese direction words used in movimento.feature.
const direction = `(cima|baixo|a esquerda|a direita)`

// Steps for movimento.feature (implemented in T12).
func (w *world) registerMovementSteps(sc *godog.ScenarioContext) {
	sc.Step(`^"([^"]*)" segura a tecla "([^"]*)"$`, pending2)
	sc.Step(`^"([^"]*)" segura as teclas "([^"]*)" e "([^"]*)"$`, pending3)
	sc.Step(`^"([^"]*)" se move para `+direction+`$`, pending2)
	sc.Step(`^"([^"]*)" vê "([^"]*)" se mover para `+direction+`$`, pending3)
	sc.Step(`^"([^"]*)" se move para cima e para a direita$`, pending1)
	sc.Step(`^a velocidade de "([^"]*)" é a mesma do movimento em linha reta$`, pending1)
	sc.Step(`^que "([^"]*)" está segurando a tecla "([^"]*)"$`, pending2)
	sc.Step(`^"([^"]*)" solta a tecla "([^"]*)"$`, pending2)
	sc.Step(`^"([^"]*)" para de se mover$`, pending1)
	sc.Step(`^que "([^"]*)" está encostado em um obstáculo à sua direita$`, pending1)
	sc.Step(`^a posição de "([^"]*)" não atravessa o obstáculo$`, pending1)
	sc.Step(`^que "([^"]*)" está na borda esquerda da Praça$`, pending1)
	sc.Step(`^"([^"]*)" continua dentro da Praça$`, pending1)
}
