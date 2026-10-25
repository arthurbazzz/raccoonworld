// Package clock provides the real wall clock for production.
package clock

import "time"

// System implements usecase.Clock with time.Now.
type System struct{}

// Now returns the current time.
func (System) Now() time.Time { return time.Now() }
