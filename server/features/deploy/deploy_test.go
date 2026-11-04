// Package deploy runs the RW-007 smoke scenarios against a running Docker Compose stack.
//
// It only runs when RW_STACK_URL points at the stack (e.g. http://localhost);
// otherwise it is skipped, so plain `go test ./...` does not need Docker.
package deploy

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/coder/websocket"
	"github.com/cucumber/godog"
)

var featurePaths = []string{"../../../specs/007-deploy-docker/features"}

func TestDeployStack(t *testing.T) {
	baseURL := os.Getenv("RW_STACK_URL")
	if baseURL == "" {
		t.Skip("RW_STACK_URL not set; start the stack and set it to run the smoke")
	}

	suite := godog.TestSuite{
		Name: "raccoons-world-deploy",
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			w := &stackWorld{baseURL: strings.TrimSuffix(baseURL, "/")}
			sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
				if w.conn != nil {
					_ = w.conn.CloseNow()
				}
				return ctx, nil
			})
			w.registerHTTPSteps(sc)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    featurePaths,
			Tags:     os.Getenv("GODOG_TAGS"),
			Strict:   true, // pending/undefined steps fail: red before green
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("godog scenarios failed")
	}
}

// stackWorld holds per-scenario state.
type stackWorld struct {
	baseURL string // public address of the stack, e.g. http://localhost

	status int    // status of the last HTTP response
	body   []byte // body of the last HTTP response

	conn *websocket.Conn // WebSocket opened through the proxy
}
