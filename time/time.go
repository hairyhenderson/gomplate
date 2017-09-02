package time

import (
	"time"
)

// ZoneName - a convenience function for determining the current timezone's name
func ZoneName() string {
	n, _ := time.Now().Zone()
	return n
}
