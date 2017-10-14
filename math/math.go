package math

// AddInt -
func AddInt(n ...int64) int64 {
	x := int64(0)
	for _, i := range n {
		x += i
	}
	return x
}

// MulInt -
func MulInt(n ...int64) int64 {
	var x int64 = 1
	for _, i := range n {
		x *= i
	}
	return x
}
