package features

import "github.com/cucumber/godog"

// Steps for conexao.feature (implemented in T14).
func (w *world) registerConnectionSteps(sc *godog.ScenarioContext) {
	sc.Step(`^a conexão de "([^"]*)" cai$`, pending1)
	sc.Step(`^em até (\d+) segundos "([^"]*)" deixa de ver "([^"]*)"$`, pending3)
}
