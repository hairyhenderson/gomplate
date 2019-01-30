package funcs

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	gotime "time"

	"github.com/hairyhenderson/gomplate/conv"
	"github.com/hairyhenderson/gomplate/time"
)

var (
	timeNS     *TimeFuncs
	timeNSInit sync.Once
)

// TimeNS -
func TimeNS() *TimeFuncs {
	timeNSInit.Do(func() {
		timeNS = &TimeFuncs{
			ANSIC:       gotime.ANSIC,
			UnixDate:    gotime.UnixDate,
			RubyDate:    gotime.RubyDate,
			RFC822:      gotime.RFC822,
			RFC822Z:     gotime.RFC822Z,
			RFC850:      gotime.RFC850,
			RFC1123:     gotime.RFC1123,
			RFC1123Z:    gotime.RFC1123Z,
			RFC3339:     gotime.RFC3339,
			RFC3339Nano: gotime.RFC3339Nano,
			Kitchen:     gotime.Kitchen,
			Stamp:       gotime.Stamp,
			StampMilli:  gotime.StampMilli,
			StampMicro:  gotime.StampMicro,
			StampNano:   gotime.StampNano,
		}
	})
	return timeNS
}

// AddTimeFuncs -
func AddTimeFuncs(f map[string]interface{}) {
	f["time"] = TimeNS
}

// TimeFuncs -
type TimeFuncs struct {
	ANSIC       string
	UnixDate    string
	RubyDate    string
	RFC822      string
	RFC822Z     string
	RFC850      string
	RFC1123     string
	RFC1123Z    string
	RFC3339     string
	RFC3339Nano string
	Kitchen     string
	Stamp       string
	StampMilli  string
	StampMicro  string
	StampNano   string
}

// ZoneName - return the local system's time zone's name
func (f *TimeFuncs) ZoneName() string {
	return time.ZoneName()
}

// ZoneOffset - return the local system's time zone's name
func (f *TimeFuncs) ZoneOffset() int {
	return time.ZoneOffset()
}

// Parse -
func (f *TimeFuncs) Parse(layout string, value interface{}) (gotime.Time, error) {
	return gotime.Parse(layout, conv.ToString(value))
}

// ParseLocal -
func (f *TimeFuncs) ParseLocal(layout string, value interface{}) (gotime.Time, error) {
	return gotime.ParseInLocation(layout, conv.ToString(value), gotime.Local)
}

// ParseInLocation -
func (f *TimeFuncs) ParseInLocation(layout, location string, value interface{}) (gotime.Time, error) {
	loc, err := gotime.LoadLocation(location)
	if err != nil {
		return gotime.Time{}, err
	}
	return gotime.ParseInLocation(layout, conv.ToString(value), loc)
}

// Now -
func (f *TimeFuncs) Now() gotime.Time {
	return gotime.Now()
}

// Unix - convert UNIX time (in seconds since the UNIX epoch) into a time.Time for further processing
// Takes a string or number (int or float)
func (f *TimeFuncs) Unix(in interface{}) (gotime.Time, error) {
	sec, nsec, err := parseNum(in)
	if err != nil {
		return gotime.Time{}, err
	}
	return gotime.Unix(sec, nsec), nil
}

// Nanosecond -
func (f *TimeFuncs) Nanosecond(n interface{}) gotime.Duration {
	return gotime.Nanosecond * gotime.Duration(conv.ToInt64(n))
}

// Microsecond -
func (f *TimeFuncs) Microsecond(n interface{}) gotime.Duration {
	return gotime.Microsecond * gotime.Duration(conv.ToInt64(n))
}

// Millisecond -
func (f *TimeFuncs) Millisecond(n interface{}) gotime.Duration {
	return gotime.Millisecond * gotime.Duration(conv.ToInt64(n))
}

// Second -
func (f *TimeFuncs) Second(n interface{}) gotime.Duration {
	return gotime.Second * gotime.Duration(conv.ToInt64(n))
}

// Minute -
func (f *TimeFuncs) Minute(n interface{}) gotime.Duration {
	return gotime.Minute * gotime.Duration(conv.ToInt64(n))
}

// Hour -
func (f *TimeFuncs) Hour(n interface{}) gotime.Duration {
	return gotime.Hour * gotime.Duration(conv.ToInt64(n))
}

// ParseDuration -
func (f *TimeFuncs) ParseDuration(n interface{}) (gotime.Duration, error) {
	return gotime.ParseDuration(conv.ToString(n))
}

// Since -
func (f *TimeFuncs) Since(n gotime.Time) gotime.Duration {
	return gotime.Since(n)
}

// Until -
func (f *TimeFuncs) Until(n gotime.Time) gotime.Duration {
	return gotime.Until(n)
}

// convert a number input to a pair of int64s, representing the integer portion and the decimal remainder
// this can handle a string as well as any integer or float type
// precision is at the "nano" level (i.e. 1e+9)
func parseNum(in interface{}) (integral int64, fractional int64, err error) {
	if s, ok := in.(string); ok {
		ss := strings.Split(s, ".")
		if len(ss) > 2 {
			return 0, 0, fmt.Errorf("can not parse '%s' as a number - too many decimal points", s)
		}
		if len(ss) == 1 {
			integral, err := strconv.ParseInt(s, 0, 64)
			return integral, 0, err
		}
		integral, err := strconv.ParseInt(ss[0], 0, 64)
		if err != nil {
			return integral, 0, err
		}
		fractional, err = strconv.ParseInt(padRight(ss[1], "0", 9), 0, 64)
		return integral, fractional, err
	}
	if s, ok := in.(fmt.Stringer); ok {
		return parseNum(s.String())
	}
	if i, ok := in.(int); ok {
		return int64(i), 0, nil
	}
	if u, ok := in.(uint64); ok {
		return int64(u), 0, nil
	}
	if f, ok := in.(float64); ok {
		return 0, 0, fmt.Errorf("can not parse floating point number (%f) - use a string instead", f)
	}
	if in == nil {
		return 0, 0, nil
	}
	return 0, 0, nil
}

// pads a number with zeroes
func padRight(in, pad string, length int) string {
	for {
		in += pad
		if len(in) > length {
			return in[0:length]
		}
	}
}
