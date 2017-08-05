package typeconv

import "strconv"

// MustParseBool - wrapper for strconv.ParseBool that returns false in the case of error
func MustParseBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

// MustParseInt - wrapper for strconv.ParseInt that returns 0 in the case of error
func MustParseInt(s string, base, bitSize int) int64 {
	i, _ := strconv.ParseInt(s, base, bitSize)
	return i
}

// MustParseFloat - wrapper for strconv.ParseFloat that returns 0 in the case of error
func MustParseFloat(s string, bitSize int) float64 {
	i, _ := strconv.ParseFloat(s, bitSize)
	return i
}

// MustParseUint - wrapper for strconv.ParseUint that returns 0 in the case of error
func MustParseUint(s string, base, bitSize int) uint64 {
	i, _ := strconv.ParseUint(s, base, bitSize)
	return i
}

// MustAtoi - wrapper for strconv.Atoi that returns 0 in the case of error
func MustAtoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
