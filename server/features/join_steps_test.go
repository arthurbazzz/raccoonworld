package features

import "github.com/cucumber/godog"

// Steps for entrar.feature (implemented in T11).
func (w *world) registerJoinSteps(sc *godog.ScenarioContext) {
	sc.Step(`^que a sala "([^"]*)" está aberta$`, pending1)
	sc.Step(`^o visitante entra com o apelido "([^"]*)"$`, pending1)
	sc.Step(`^"([^"]*)" aparece na Praça$`, pending1)
	sc.Step(`^"([^"]*)" recebe a lista dos jogadores já presentes$`, pending1)
	sc.Step(`^que "([^"]*)" está na Praça$`, pending1)
	sc.Step(`^que "([^"]*)" e "([^"]*)" estão na Praça$`, pending2)
	sc.Step(`^outro visitante tenta entrar com o apelido "([^"]*)"$`, pending1)
	sc.Step(`^o visitante tenta entrar com o apelido "([^"]*)"$`, pending1)
	sc.Step(`^a entrada é recusada com o motivo "([^"]*)"$`, pending1)
	sc.Step(`^que a Praça está com a capacidade máxima$`, pending0)
	sc.Step(`^"([^"]*)" entra na Praça$`, pending1)
	sc.Step(`^"([^"]*)" é avisado de que "([^"]*)" entrou$`, pending2)
}
