// Package features runs the RW-001 Gherkin scenarios against the use cases.
//
// The .feature files live in specs/ and are the single source of truth;
// both godog (here) and playwright-bdd (web/) read them from there.
package features

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
)

// featurePaths are relative to this package directory (go test runs here).
var featurePaths = []string{
	"../../specs/001-sala-inicial/features",
}

func TestFeatures(t *testing.T) {
	tags := "~@ui" // browser-only scenarios run in playwright-bdd
	if extra := os.Getenv("GODOG_TAGS"); extra != "" {
		tags += " && " + extra
	}

	suite := godog.TestSuite{
		Name:                "rw-001",
		ScenarioInitializer: initializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    featurePaths,
			Tags:     tags,
			Strict:   true, // pending/undefined steps fail: red before green
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("godog scenarios failed")
	}
}

// world holds per-scenario state. A fresh one is built for every scenario.
type world struct {
	clock       *fakeClock
	broadcaster *fakeBroadcaster
}

func initializeScenario(sc *godog.ScenarioContext) {
	w := &world{
		clock:       newFakeClock(),
		broadcaster: &fakeBroadcaster{},
	}
	w.registerJoinSteps(sc)
	w.registerMovementSteps(sc)
	w.registerChatSteps(sc)
	w.registerConnectionSteps(sc)
}
