// Package random contains functions for generating random values
package random

import (
	"fmt"
	"math"
	"math/rand/v2"
	"regexp"
	"slices"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/puzpuzpuz/xsync/v4"
)

// Default set, matches "[a-zA-Z0-9_.-]"
const defaultSet = "-.0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"

// graphicRunes returns every graphic rune in order, which are the candidates
// that matchChars tests the pattern against. This is the same as
// filterRange(0, unicode.MaxRune), but reading the Unicode tables directly is
// much faster than testing every rune.
var graphicRunes = sync.OnceValue(func() []rune {
	out := []rune{}
	for _, t := range unicode.GraphicRanges {
		for _, r := range t.R16 {
			for c := rune(r.Lo); c <= rune(r.Hi); c += rune(r.Stride) {
				out = append(out, c)
			}
		}
		for _, r := range t.R32 {
			for c := rune(r.Lo); c <= rune(r.Hi); c += rune(r.Stride) {
				out = append(out, c)
			}
		}
	}
	slices.Sort(out)

	return out
})

// matchCache holds the result of matchChars for each pattern, since testing
// a pattern against every graphic rune is slow. The cached slices are never
// modified.
var matchCache = xsync.NewMap[string, []rune]()

// StringRE - Generate a random string that matches a given regular
// expression. Defaults to "[a-zA-Z0-9_.-]"
func StringRE(count int, match string) (r string, err error) {
	chars := []rune(defaultSet)
	if match != "" {
		chars, err = cachedMatchChars(match)
		if err != nil {
			return "", err
		}

		if len(chars) == 0 {
			return "", fmt.Errorf("no printable characters match %q", match)
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
	if count < 0 {
		return "", fmt.Errorf("count must not be negative (was %d)", count)
	}

	s := make([]rune, count)
	for i := range s {
		//nolint:gosec
		s[i] = chars[rand.IntN(len(chars))]
	}
	return string(s), nil
}

func filterRange(lower, upper rune) []rune {
	// there are no characters outside this range, and stopping at
	// unicode.MaxRune keeps r from overflowing when upper is math.MaxInt32
	lower = max(lower, 0)
	upper = min(upper, unicode.MaxRune)

	out := []rune{}
	for r := lower; r <= upper; r++ {
		if unicode.IsGraphic(r) {
			out = append(out, r)
		}
	}
	return out
}

func cachedMatchChars(match string) ([]rune, error) {
	if chars, ok := matchCache.Load(match); ok {
		return chars, nil
	}

	chars, err := matchChars(match)
	if err != nil {
		return nil, err
	}

	matchCache.Store(match, chars)

	return chars, nil
}

func matchChars(match string) ([]rune, error) {
	r, err := regexp.Compile(match)
	if err != nil {
		return nil, err
	}

	out := []rune{}
	buf := make([]byte, 0, utf8.UTFMax)
	for _, c := range graphicRunes() {
		buf = utf8.AppendRune(buf[:0], c)
		if r.Match(buf) {
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
