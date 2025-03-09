// Package random contains functions for generating random values
package random

import (
	"fmt"
	"math"
	"math/rand/v2"
	"regexp"
	"unicode"
)

// Default set, matches "[a-zA-Z0-9_.-]"
const defaultSet = "-.0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"

// StringRE - Generate a random string that matches a given regular
// expression. Defaults to "[a-zA-Z0-9_.-]"
func StringRE(count int, match string) (r string, err error) {
	chars := []rune(defaultSet)
	if match != "" {
		chars, err = matchChars(match)
		if err != nil {
			return "", err
		}
	}

	return rndString(count, chars)
}

// StringBounds returns a random string of characters with a codepoint
// between the lower and upper bounds. Only valid characters are returned
// and if a range is given where no valid characters can be found, an error
// will be returned.
func StringBounds(count int, lower, upper rune) (r string, err error) {
	chars := filterRange(lower, upper)
	if len(chars) == 0 {
		return "", fmt.Errorf("no printable codepoints found between U%#q and U%#q", lower, upper)
	}
	return rndString(count, chars)
}

// produce a string containing a random selection of given characters
func rndString(count int, chars []rune) (string, error) {
	s := make([]rune, count)
	for i := range s {
		//nolint:gosec
		s[i] = chars[rand.IntN(len(chars))]
	}
	return string(s), nil
}

func filterRange(lower, upper rune) []rune {
	out := []rune{}
	for r := lower; r <= upper; r++ {
		if unicode.IsGraphic(r) {
			out = append(out, r)
		}
	}
	return out
}

func matchChars(match string) ([]rune, error) {
	r, err := regexp.Compile(match)
	if err != nil {
		return nil, err
	}
	candidates := filterRange(0, unicode.MaxRune)
	out := []rune{}
	for _, c := range candidates {
		if r.MatchString(string(c)) {
			out = append(out, c)
		}
	}
	return out, nil
}

// Item -
func Item(items []any) (any, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("expected a non-empty array or slice")
	}
	if len(items) == 1 {
		return items[0], nil
	}

	//nolint:gosec
	n := rand.IntN(len(items))
	return items[n], nil
}

// Number -
//
//nolint:revive
func Number(min, max int64) (int64, error) {
	if min > max {
		return 0, fmt.Errorf("min must not be greater than max (was %d, %d)", min, max)
	}
	if min == math.MinInt64 {
		min++
	}
	if max-min >= (math.MaxInt64 >> 1) {
		return 0, fmt.Errorf("spread between min and max too high - must not be greater than 63-bit maximum (%d - %d = %d)", max, min, max-min)
	}

	//nolint:gosec
	return rand.Int64N(max-min+1) + min, nil
}

// Float - For now this is really just a wrapper around `rand.Float64`
//
//nolint:revive
func Float(min, max float64) (float64, error) {
	//nolint:gosec
	return min + rand.Float64()*(max-min), nil
}
