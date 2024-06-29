---
title: time functions
menu:
  main:
    parent: functions
---

This namespace wraps Go's [`time` package](https://pkg.go.dev/time/), and a
few of the functions return a `time.Time` value. All of the
[`time.Time` functions](https://pkg.go.dev/time/#Time) can then be used to
convert, adjust, or format the time in your template.

### Reference time

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

Some operations (such as [`Time.Add`](https://pkg.go.dev/time/#Time.Add) and
[`Time.Round`](https://pkg.go.dev/time/#Time.Round)) require a
[`Duration`](https://pkg.go.dev/time/#Duration) value. These can be created
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

For other durations, such as `2h10m`, [`time.ParseDuration`](#timeparseduration) can be used.

## `time.Now`

Returns the current local time, as a `time.Time`. This wraps [`time.Now`](https://pkg.go.dev/time/#Now).

Usually, further functions are called using the value returned by `Now`.

_Added in gomplate [v2.1.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.1.0)_
### Usage

```
time.Now
```


### Examples

Usage with [`UTC`](https://pkg.go.dev/time/#Time.UTC) and [`Format`](https://pkg.go.dev/time/#Time.Format):
```console
$ gomplate -i '{{ (time.Now).UTC.Format "Day 2 of month 1 in year 2006 (timezone MST)" }}'
Day 14 of month 10 in year 2017 (timezone UTC)
```
Usage with [`AddDate`](https://pkg.go.dev/time/#Time.AddDate):
```console
$ date
Sat Oct 14 09:57:02 EDT 2017
$ gomplate -i '{{ ((time.Now).AddDate 0 1 0).Format "Mon Jan 2 15:04:05 MST 2006" }}'
Tue Nov 14 09:57:02 EST 2017
```

_(notice how the TZ adjusted for daylight savings!)_
Usage with [`IsDST`](https://pkg.go.dev/time/#Time.IsDST):
```console
$ gomplate -i '{{ $t := time.Now }}At the tone, the time will be {{ ($t.Round (time.Minute 1)).Add (time.Minute 1) }}.
  It is{{ if not $t.IsDST }} not{{ end }} daylight savings time.
  ... ... BEEP'
At the tone, the time will be 2022-02-10 09:01:00 -0500 EST.
It is not daylight savings time.
... ... BEEP
```

## `time.Parse`

Parses a timestamp defined by the given layout. This wraps [`time.Parse`](https://pkg.go.dev/time/#Parse).

A number of pre-defined layouts are provided as constants, defined
[here](https://pkg.go.dev/time/#pkg-constants).

Just like [`time.Now`](#timenow), this is usually used in conjunction with
other functions.

_Note: In the absence of a time zone indicator, `time.Parse` returns a time in UTC._

_Added in gomplate [v2.1.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.1.0)_
### Usage

```
time.Parse layout timestamp
```
```
timestamp | time.Parse layout
```

### Arguments

| name | description |
|------|-------------|
| `layout` | _(required)_ The layout string to parse with |
| `timestamp` | _(required)_ The timestamp to parse |

### Examples

Usage with [`Format`](https://pkg.go.dev/time/#Time.Format):
```console
$ gomplate -i '{{ (time.Parse "2006-01-02" "1993-10-23").Format "Monday January 2, 2006 MST" }}'
Saturday October 23, 1993 UTC
```

## `time.ParseDuration`

Parses a duration string. This wraps [`time.ParseDuration`](https://pkg.go.dev/time/#ParseDuration).

A duration string is a possibly signed sequence of decimal numbers, each with
optional fraction and a unit suffix, such as `300ms`, `-1.5h` or `2h45m`. Valid
time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

_Added in gomplate [v2.1.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.1.0)_
### Usage

```
time.ParseDuration duration
```
```
duration | time.ParseDuration
```

### Arguments

| name | description |
|------|-------------|
| `duration` | _(required)_ The duration string to parse |

### Examples

```console
$ gomplate -i '{{ (time.Now).Format time.Kitchen }}
{{ ((time.Now).Add (time.ParseDuration "2h30m")).Format time.Kitchen }}'
12:43AM
3:13AM
```

## `time.ParseLocal`

Same as [`time.Parse`](#timeparse), except that in the absence of a time zone
indicator, the timestamp wil be parsed in the local timezone.

_Added in gomplate [v2.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.2.0)_
### Usage

```
time.ParseLocal layout timestamp
```
```
timestamp | time.ParseLocal layout
```

### Arguments

| name | description |
|------|-------------|
| `layout` | _(required)_ The layout string to parse with |
| `timestamp` | _(required)_ The timestamp to parse |

### Examples

Usage with [`Format`](https://pkg.go.dev/time/#Time.Format):
```console
$ bin/gomplate -i '{{ (time.ParseLocal time.Kitchen "6:00AM").Format "15:04 MST" }}'
06:00 EST
```

## `time.ParseInLocation`

Same as [`time.Parse`](#timeparse), except that the time is parsed in the given location's time zone.

This wraps [`time.ParseInLocation`](https://pkg.go.dev/time/#ParseInLocation).

_Added in gomplate [v2.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.2.0)_
### Usage

```
time.ParseInLocation layout location timestamp
```
```
timestamp | time.ParseInLocation layout location
```

### Arguments

| name | description |
|------|-------------|
| `layout` | _(required)_ The layout string to parse with |
| `location` | _(required)_ The location to parse in |
| `timestamp` | _(required)_ The timestamp to parse |

### Examples

Usage with [`Format`](https://pkg.go.dev/time/#Time.Format):
```console
$ gomplate -i '{{ (time.ParseInLocation time.Kitchen "Africa/Luanda" "6:00AM").Format "15:04 MST" }}'
06:00 LMT
```

## `time.Since`

Returns the time elapsed since a given time. This wraps [`time.Since`](https://pkg.go.dev/time/#Since).

It is shorthand for `time.Now.Sub t`.

_Added in gomplate [v2.5.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.5.0)_
### Usage

```
time.Since t
```
```
t | time.Since
```

### Arguments

| name | description |
|------|-------------|
| `t` | _(required)_ the `Time` to calculate since |

### Examples

```console
$ gomplate -i '{{ $t := time.Parse time.RFC3339 "1970-01-01T00:00:00Z" }}time since the epoch:{{ time.Since $t }}'
time since the epoch:423365h0m24.353828924s
```

## `time.Unix`

Returns the local `Time` corresponding to the given Unix time, in seconds since
January 1, 1970 UTC. Note that fractional seconds can be used to denote
milliseconds, but must be specified as a string, not a floating point number.

_Added in gomplate [v2.1.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.1.0)_
### Usage

```
time.Unix time
```
```
time | time.Unix
```

### Arguments

| name | description |
|------|-------------|
| `time` | _(required)_ the time to parse |

### Examples

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

## `time.Until`

Returns the duration until a given time. This wraps [`time.Until`](https://pkg.go.dev/time/#Until).

It is shorthand for `$t.Sub time.Now`.

_Added in gomplate [v2.5.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.5.0)_
### Usage

```
time.Until t
```
```
t | time.Until
```

### Arguments

| name | description |
|------|-------------|
| `t` | _(required)_ the `Time` to calculate until |

### Examples

```console
$ gomplate -i '{{ $t := time.Parse time.RFC3339 "2020-01-01T00:00:00Z" }}only {{ time.Until $t }} to go...'
only 14922h56m46.578625891s to go...
```

Or, less precise:
```console
$ bin/gomplate -i '{{ $t := time.Parse time.RFC3339 "2020-01-01T00:00:00Z" }}only {{ (time.Until $t).Round (time.Hour 1) }} to go...'
only 14923h0m0s to go...
```

## `time.ZoneName`

Return the local system's time zone's name.

_Added in gomplate [v2.1.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.1.0)_
### Usage

```
time.ZoneName
```


### Examples

```console
$ gomplate -i '{{time.ZoneName}}'
EDT
```

## `time.ZoneOffset`

Return the local system's time zone offset, in seconds east of UTC.

_Added in gomplate [v2.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.2.0)_
### Usage

```
time.ZoneOffset
```


### Examples

```console
$ gomplate -i '{{time.ZoneOffset}}'
-14400
```
