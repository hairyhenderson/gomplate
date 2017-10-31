package time

import (
	"time"
)

// ZoneName - a convenience function for determining the current timezone's name
func ZoneName() string {
	n, _ := time.Now().Zone()
	return n
}

// ZoneOffset - determine the current timezone's offset, in seconds east of UTC
func ZoneOffset() int {
	_, o := time.Now().Zone()
	return o
}
