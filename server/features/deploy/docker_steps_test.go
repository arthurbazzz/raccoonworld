package deploy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

const (
	// defaultProject is the compose project name set in deploy/compose.yaml.
	defaultProject = "raccoonworld"
	dockerTimeout  = 2 * time.Minute
	restartWait    = 90 * time.Second
)

// registerDockerSteps binds the steps that inspect the stack with the docker
// CLI: published ports and Postgres data. They need the stack on this machine.
func (w *stackWorld) registerDockerSteps(sc *godog.ScenarioContext) {
	sc.Step(`^a única porta publicada pela stack é a (\d+)$`, w.onlyPublishedPort)
	sc.Step(`^o Postgres não aceita conexão de fora da stack$`, w.postgresNotPublished)
	sc.Step(`^que existe a tabela "([^"]*)" com uma linha no Postgres$`, w.createProbeTable)
	sc.Step(`^a stack é derrubada e sobe de novo$`, w.restartStack)
	sc.Step(`^a tabela "([^"]*)" ainda tem a linha$`, w.probeTableHasRow)
}

// project is the compose project to inspect. Env: RW_STACK_PROJECT.
func project() string {
	if p := os.Getenv("RW_STACK_PROJECT"); p != "" {
		return p
	}
	return defaultProject
}

// compose runs `docker compose -p <project> args...` and returns stdout.
func compose(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dockerTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", append([]string{"compose", "-p", project()}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker compose %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String(), nil
}

// psql runs SQL inside the postgres container, with the credentials it already has.
func psql(sql string) (string, error) {
	return compose("exec", "-T", "postgres", "sh", "-c",
		`psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tAc "$0"`, sql)
}

type composeService struct {
	Service    string
	Publishers []struct {
		PublishedPort int
	}
}

// services lists the running services and their published ports.
func services() ([]composeService, error) {
	out, err := compose("ps", "--format", "json")
	if err != nil {
		return nil, err
	}
	var list []composeService
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var s composeService
		if err := json.Unmarshal([]byte(line), &s); err != nil {
			return nil, fmt.Errorf("docker compose ps: %w", err)
		}
		list = append(list, s)
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("no services running in project %q", project())
	}
	return list, sc.Err()
}

// publishedPorts returns the host ports a service publishes (0 means not published).
func (s composeService) publishedPorts() []int {
	var ports []int
	for _, p := range s.Publishers {
		if p.PublishedPort != 0 {
			ports = append(ports, p.PublishedPort)
		}
	}
	return ports
}

func (w *stackWorld) onlyPublishedPort(port int) error {
	list, err := services()
	if err != nil {
		return err
	}
	for _, s := range list {
		for _, p := range s.publishedPorts() {
			if p != port {
				return fmt.Errorf("service %q publishes port %d, want only %d", s.Service, p, port)
			}
		}
	}
	return nil
}

func (w *stackWorld) postgresNotPublished() error {
	list, err := services()
	if err != nil {
		return err
	}
	for _, s := range list {
		if s.Service == "postgres" {
			if ports := s.publishedPorts(); len(ports) > 0 {
				return fmt.Errorf("postgres publishes ports %v, want none", ports)
			}
			return nil
		}
	}
	return errors.New("postgres service not found")
}

func (w *stackWorld) createProbeTable(table string) error {
	w.probeTable = table
	_, err := psql(fmt.Sprintf(
		"DROP TABLE IF EXISTS %[1]s; CREATE TABLE %[1]s (note text); INSERT INTO %[1]s VALUES ('rw-007');", table))
	return err
}

// restartStack stops every container and starts them again, keeping the volumes.
func (w *stackWorld) restartStack() error {
	if _, err := compose("stop"); err != nil {
		return err
	}
	if _, err := compose("start"); err != nil {
		return err
	}
	deadline := time.Now().Add(restartWait)
	for {
		_, dbErr := psql("SELECT 1")
		healthErr := w.stackIsUp()
		if dbErr == nil && healthErr == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("stack not back after restart: %w", errors.Join(dbErr, healthErr))
		}
		time.Sleep(time.Second)
	}
}

func (w *stackWorld) probeTableHasRow(table string) error {
	out, err := psql(fmt.Sprintf("SELECT note FROM %s;", table))
	if err != nil {
		return err
	}
	if got := strings.TrimSpace(out); got != "rw-007" {
		return fmt.Errorf("table %q holds %q, want the row written before the restart", table, got)
	}
	return nil
}

// dropProbeTable removes the scenario's table, so the stack is left as found.
func (w *stackWorld) dropProbeTable() {
	if w.probeTable != "" {
		_, _ = psql("DROP TABLE IF EXISTS " + w.probeTable + ";")
	}
}
