// Package features runs the RW-001 Gherkin scenarios against the use cases.
//
// The .feature files live in specs/ and are the single source of truth;
// both godog (here) and playwright-bdd (web/) read them from there.
package features

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"raccoons-world/server/internal/domain/room"
	"raccoons-world/server/internal/usecase"
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

	runner *usecase.RoomRunner
	ticks  chan time.Time
	stop   context.CancelFunc
	done   chan struct{}

	join      *usecase.JoinRoom
	setInput  *usecase.SetInput
	sendChat  *usecase.SendChat
	keepAlive *usecase.KeepAlive

	dropped map[string]bool // nickname -> connection dropped (no more heartbeats)

	lastChatText string // text of the last chat attempt
	lastChatErr  error  // error of the last chat attempt

	held        map[string]room.Input // nickname -> keys being held
	mark        map[string]room.Vec   // nickname -> position when keys last changed
	nextVisitor int
	ids         map[string]room.PlayerID // nickname -> id of players who joined
	lastAttempt room.PlayerID            // id used by the last join attempt
	lastErr     error                    // error of the last join attempt
}

func initializeScenario(sc *godog.ScenarioContext) {
	w := &world{
		clock:       newFakeClock(),
		broadcaster: &fakeBroadcaster{},
		ids:         map[string]room.PlayerID{},
		held:        map[string]room.Input{},
		mark:        map[string]room.Vec{},
		dropped:     map[string]bool{},
	}
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		w.closeRoom()
		return ctx, nil // the step error is already reported by godog
	})
	w.registerJoinSteps(sc)
	w.registerMovementSteps(sc)
	w.registerChatSteps(sc)
	w.registerConnectionSteps(sc)
}
