package features

import "github.com/cucumber/godog"

// Placeholders for steps not implemented yet. With Strict mode they keep
// the scenario red until a task replaces them.

func pending1(_ string) error       { return godog.ErrPending }
func pending3(_, _, _ string) error { return godog.ErrPending }
