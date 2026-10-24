// Package math contains set of basic math functions to be able to perform simple arithmetic operations
package math

import (
	"fmt"
	"math"
)

// AddInt -
//
// Deprecated: this function is unused and will be removed in a future
// release. Sum the values directly instead.
func AddInt(n ...int64) int64 {
	x := int64(0)
	for _, i := range n {
		x += i
	}
	return x
}

// MulInt -
//
// Deprecated: this function is unused and will be removed in a future
// release. Multiply the values directly instead.
func MulInt(n ...int64) int64 {
	var x int64 = 1
	for _, i := range n {
		x *= i
	}
	return x
}

// Seq - return a sequence from `start` to `end`, in steps of `step`. The
// sequence always goes from start towards end, whatever the sign of step. It
// panics if the sequence is too long to allocate.
func Seq(start, end, step int64) []int64 {
	// a step of 0 just returns an empty sequence
	if step == 0 {
		return []int64{}
	}

	// The distance between two int64s doesn't always fit in an int64, but it
	// always fits in a uint64, so count the steps in uint64 to avoid overflow.
	//nolint:gosec // G115 - the two's complement conversion is intended
	stride := uint64(step)
	if step < 0 {
		stride = -stride
	}

	//nolint:gosec // G115 - as above
	dist := uint64(end) - uint64(start)
	if end < start {
		dist = -dist
	}

	n := dist / stride
	if n >= math.MaxInt {
		panic(fmt.Sprintf("sequence from %d to %d in steps of %d is too long", start, end, step))
	}

	// Every value in the sequence fits in an int64, so adding with wraparound
	// gives the right values even when the stride itself doesn't fit.
	//nolint:gosec // G115 - as above
	delta := int64(stride)
	if end < start {
		delta = -delta
	}

	seq := make([]int64, n+1)
	v := start
	for i := range seq {
		seq[i] = v
		v += delta
	}

	return seq
}
