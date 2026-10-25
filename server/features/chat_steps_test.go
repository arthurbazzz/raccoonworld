package features

import "github.com/cucumber/godog"

// Steps for chat.feature (implemented in T13).
func (w *world) registerChatSteps(sc *godog.ScenarioContext) {
	sc.Step(`^"([^"]*)" envia a mensagem "([^"]*)"$`, pending2)
	sc.Step(`^"([^"]*)" vê o balão "([^"]*)" sobre "([^"]*)"$`, pending3)
	sc.Step(`^"([^"]*)" acabou de enviar uma mensagem$`, pending1)
	sc.Step(`^"([^"]*)" envia outra mensagem em menos de 1 segundo$`, pending1)
	sc.Step(`^a segunda mensagem é descartada$`, pending0)
	sc.Step(`^"([^"]*)" envia uma mensagem com (\d+) caracteres$`, pending2)
	sc.Step(`^a mensagem é descartada$`, pending0)
}
