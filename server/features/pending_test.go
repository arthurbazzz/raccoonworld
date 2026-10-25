package features

import "github.com/cucumber/godog"

// Placeholders for steps not implemented yet. Each task from T11 onward
// replaces them with real implementations; with Strict mode they keep the
// scenario red until then.

func pending0() error               { return godog.ErrPending }
func pending1(_ string) error       { return godog.ErrPending }
func pending2(_, _ string) error    { return godog.ErrPending }
func pending3(_, _, _ string) error { return godog.ErrPending }
