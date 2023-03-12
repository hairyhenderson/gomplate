// Package time contains functions to help work with date and time
package time

import (
	"time"

	"github.com/hairyhenderson/gomplate/v4/env"
)

// ZoneName - a convenience function for determining the current timezone's name
func ZoneName() string {
	n, _ := zone()
	return n
}

// ZoneOffset - determine the current timezone's offset, in seconds east of UTC
func ZoneOffset() int {
	_, o := zone()
	return o
}

func zone() (string, int) {
	// re-read TZ env var in case it's changed since the process started.
	// This may happen in certain rare instances when this is being called as a
	// library, or in a test. It allows for a bit more flexibility too, as
	// changing time.Local is prone to data races.
	tz := env.Getenv("TZ", "Local")
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.Local
	}
	return time.Now().In(loc).Zone()
}
