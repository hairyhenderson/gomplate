---
title: time functions
menu:
  main:
    parent: functions
---

This namespace wraps Go's [`time` package](https://golang.org/pkg/time/), and a
few of the functions return a `time.Time` value. All of the 
[`time.Time` functions](https://golang.org/pkg/time/#Time) can then be used to
convert, adjust, or format the time in your template.

An important difference between this and many other time/date utilities is how
parsing and formatting is accomplished. Instead of relying solely on pre-defined
formats, or having a complex system of variables, formatting is accomplished by
declaring an example of the layout you wish to display.

This uses a _reference time_, which is:

```
Mon Jan 2 15:04:05 -0700 MST 2006
```

### Constants

#### format layouts 

Some pre-defined layouts have been provided for convenience:

| layout name | value |
|-------------|-------|
| `time.ANSIC`       | `"Mon Jan _2 15:04:05 2006"` |
| `time.UnixDate`    | `"Mon Jan _2 15:04:05 MST 2006"` |
| `time.RubyDate`    | `"Mon Jan 02 15:04:05 -0700 2006"` |
| `time.RFC822`      | `"02 Jan 06 15:04 MST"` |
| `time.RFC822Z`     | `"02 Jan 06 15:04 -0700"` // RFC822 with numeric zone |
| `time.RFC850`      | `"Monday, 02-Jan-06 15:04:05 MST"` |
| `time.RFC1123`     | `"Mon, 02 Jan 2006 15:04:05 MST"` |
| `time.RFC1123Z`    | `"Mon, 02 Jan 2006 15:04:05 -0700"` // RFC1123 with numeric zone |
| `time.RFC3339`     | `"2006-01-02T15:04:05Z07:00"` |
| `time.RFC3339Nano` | `"2006-01-02T15:04:05.999999999Z07:00"` |
| `time.Kitchen`     | `"3:04PM"` |
| `time.Stamp`      | `"Jan _2 15:04:05"` |
| `time.StampMilli` | `"Jan _2 15:04:05.000" `|
| `time.StampMicro` | `"Jan _2 15:04:05.000000"` |
| `time.StampNano`  | `"Jan _2 15:04:05.000000000"` |

See below for examples of how these layouts can be used.

#### durations

Some operations (such as [`Time.Add`](https://golang.org/pkg/time/#Time.Add) and 
[`Time.Round`](https://golang.org/pkg/time/#Time.Round)) require a
[`Duration`](https://golang.org/pkg/time/#Duration) value. These can be created
conveniently with the following functions:

- `time.Nanosecond`
- `time.Microsecond`
- `time.Millisecond`
- `time.Second`
- `time.Minute`
- `time.Hour`

For example:

```console
$ gomplate -i '{{ (time.Now).Format time.Kitchen }}
{{ ((time.Now).Add (time.Hour 2)).Format time.Kitchen }}'
9:05AM
11:05AM
```

## `time.Now`

Returns the current local time, as a `time.Time`. This wraps [`time.Now`](https://golang.org/pkg/time/#Now).

Usually, further functions are called using the value returned by `Now`.

### Usage
```go
time.Now
```

### Examples

Usage with [`UTC`](https://golang.org/pkg/time/#Time.UTC) and [`Format`](https://golang.org/pkg/time/#Time.Format):
```console
$ gomplate -i '{{ (time.Now).UTC.Format "Day 2 of month 1 in year 2006 (timezone MST)" }}'
Day 14 of month 10 in year 2017 (timezone UTC)
```

Usage with [`AddDate`](https://golang.org/pkg/time/#Time.AddDate):
```console
$ date
Sat Oct 14 09:57:02 EDT 2017
$ gomplate -i '{{ ((time.Now).AddDate 0 1 0).Format "Mon Jan 2 15:04:05 MST 2006" }}'
Tue Nov 14 09:57:02 EST 2017
```

_(notice how the TZ adjusted for daylight savings!)_

## `time.Parse`

Parses a timestamp defined by the given layout. This wraps [`time.Parse`](https://golang.org/pkg/time/#Parse).

A number of pre-defined layouts are provided as constants, defined
[here](https://golang.org/pkg/time/#pkg-constants).

Just like [`time.Now`](#time-now), this is usually used in conjunction with
other functions.

### Usage
```go
time.Parse layout timestamp
```

### Examples

Usage with [`Format`](https://golang.org/pkg/time/#Time.Format):
```console
$ gomplate -i '{{ (time.Parse "2006-01-02" "1993-10-23").Format "Monday January 2, 2006" }}'
Saturday October 23, 1993
```

## `time.Unix`

Returns the local `Time` corresponding to the given Unix time, in seconds since
January 1, 1970 UTC. Note that fractional seconds can be used to denote
milliseconds, but must be specified as a string, not a floating point number.

### Usage
```go
time.Unix time
```

### Example

_with whole seconds:_
```console
$ gomplate -i '{{ (time.Unix 42).UTC.Format time.Stamp}}'
Jan  1, 00:00:42
```

_with fractional seconds:_
```console
$ gomplate -i '{{ (time.Unix "123456.789").UTC.Format time.StampMilli}}'
Jan  2 10:17:36.789
```

## `time.ZoneName`

Return the local system's time zone's name.

### Usage
```go
time.ZoneName
```

### Example

```console
$ gomplate -i '{{time.ZoneName}}'
EDT
```
