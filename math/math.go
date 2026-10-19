// Package math contains set of basic math functions to be able to perform simple arithmetic operations
package math

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

// Seq - return a sequence from `start` to `end`, in steps of `step`.
func Seq(start, end, step int64) []int64 {
	// a step of 0 just returns an empty sequence
	if step == 0 {
		return []int64{}
	}

	// handle cases where step has wrong sign
	if end < start && step > 0 {
		step = -step
	}
	if end > start && step < 0 {
		step = -step
	}

	// adjust the end so it aligns exactly (avoids infinite loop!)
	end -= (end - start) % step

	seq := []int64{start}
	last := start
	for last != end {
		last = seq[len(seq)-1] + step
		seq = append(seq, last)
	}
	return seq
}
