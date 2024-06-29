---
title: random functions
menu:
  main:
    parent: functions
---

Functions for generating random values.

### About randomness

`gomplate` uses Go's [`math/rand`](https://pkg.go.dev/math/rand/) package
to generate pseudo-random numbers. Note that these functions are not suitable
for use in security-sensitive applications, such as cryptography. However,
these functions will not deplete system entropy.

## `random.ASCII`

Generates a random string of a desired length, containing the set of
printable characters from the 7-bit [ASCII](https://en.wikipedia.org/wiki/ASCII)
set. This includes _space_ (' '), but no other whitespace characters.

_Added in gomplate [v3.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.4.0)_
### Usage

```
random.ASCII count
```

### Arguments

| name | description |
|------|-------------|
| `count` | _(required)_ the length of the string to produce (number of characters) |

### Examples

```console
$ gomplate -i '{{ random.ASCII 8 }}'
_woJ%D&K
```

## `random.Alpha`

Generates a random alphabetical (`A-Z`, `a-z`) string of a desired length.

_Added in gomplate [v3.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.4.0)_
### Usage

```
random.Alpha count
```

### Arguments

| name | description |
|------|-------------|
| `count` | _(required)_ the length of the string to produce (number of characters) |

### Examples

```console
$ gomplate -i '{{ random.Alpha 42 }}'
oAqHKxHiytYicMxTMGHnUnAfltPVZDhFkVkgDvatJK
```

## `random.AlphaNum`

Generates a random alphanumeric (`0-9`, `A-Z`, `a-z`) string of a desired length.

_Added in gomplate [v3.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.4.0)_
### Usage

```
random.AlphaNum count
```

### Arguments

| name | description |
|------|-------------|
| `count` | _(required)_ the length of the string to produce (number of characters) |

### Examples

```console
$ gomplate -i '{{ random.AlphaNum 16 }}'
4olRl9mRmVp1nqSm
```

## `random.String`

Generates a random string of a desired length.

By default, the possible characters are those represented by the
regular expression `[a-zA-Z0-9_.-]` (alphanumeric, plus `_`, `.`, and `-`).

A different set of characters can be specified with a regular expression,
or by giving a range of possible characters by specifying the lower and
upper bounds. Lower/upper bounds can be specified as characters (e.g.
`"q"`, or escape sequences such as `"\U0001f0AF"`), or numeric Unicode
code-points (e.g. `48` or `0x30` for the character `0`).

When given a range of Unicode code-points, `random.String` will discard
non-printable characters from the selection. This may result in a much
smaller set of possible characters than intended, so check
the [Unicode character code charts](http://www.unicode.org/charts/) to
verify the correct code-points.

_Added in gomplate [v3.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.4.0)_
### Usage

```
random.String count [regex] [lower] [upper]
```

### Arguments

| name | description |
|------|-------------|
| `count` | _(required)_ the length of the string to produce (number of characters) |
| `regex` | _(optional)_ the regular expression that each character must match (defaults to `[a-zA-Z0-9_.-]`) |
| `lower` | _(optional)_ lower bound for a range of characters (number or single character) |
| `upper` | _(optional)_ upper bound for a range of characters (number or single character) |

### Examples

```console
$ gomplate -i '{{ random.String 8 }}'
FODZ01u_
```
```console
$ gomplate -i '{{ random.String 16 `[[:xdigit:]]` }}'
B9e0527C3e45E1f3
```
```console
$ gomplate -i '{{ random.String 20 `[\p{Canadian_Aboriginal}]` }}'
ᗄᖖᣡᕔᕫᗝᖴᒙᗌᘔᓰᖫᗵᐕᗵᙔᗠᓅᕎᔹ
```
```console
$ gomplate -i '{{ random.String 8 "c" "m" }}'
ffmidgjc
```
```console
$ gomplate -i 'You rolled... {{ random.String 3 "⚀" "⚅" }}'
You rolled... ⚅⚂⚁
```
```console
$ gomplate -i 'Poker time! {{ random.String 5 "\U0001f0a1" "\U0001f0de" }}'
Poker time! 🂼🂺🂳🃅🂪
```

## `random.Item`

Pick an element at a random from a given slice or array.

_Added in gomplate [v3.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.4.0)_
### Usage

```
random.Item items
```
```
items | random.Item
```

### Arguments

| name | description |
|------|-------------|
| `items` | _(required)_ the input array |

### Examples

```console
$ gomplate -i '{{ random.Item (seq 0 5) }}'
4
```
```console
$ export SLICE='["red", "green", "blue"]'
$ gomplate -i '{{ getenv "SLICE" | jsonArray | random.Item }}'
blue
```

## `random.Number`

Pick a random integer. By default, a number between `0` and `100`
(inclusive) is chosen, but this range can be overridden.

Note that the difference between `min` and `max` can not be larger than a
63-bit integer (i.e. the unsigned portion of a 64-bit signed integer).
The result is given as an `int64`.

_Added in gomplate [v3.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.4.0)_
### Usage

```
random.Number [min] [max]
```

### Arguments

| name | description |
|------|-------------|
| `min` | _(optional)_ The minimum value, defaults to `0`. Must be less than `max`. |
| `max` | _(optional)_ The maximum value, defaults to `100` (if no args provided) |

### Examples

```console
$ gomplate -i '{{ random.Number }}'
55
```
```console
$ gomplate -i '{{ random.Number -10 10 }}'
-3
```
```console
$ gomplate -i '{{ random.Number 5 }}'
2
```

## `random.Float`

Pick a random decimal floating-point number. By default, a number between
`0.0` and `1.0` (_exclusive_, i.e. `[0.0,1.0)`) is chosen, but this range
can be overridden.

The result is given as a `float64`.

_Added in gomplate [v3.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.4.0)_
### Usage

```
random.Float [min] [max]
```

### Arguments

| name | description |
|------|-------------|
| `min` | _(optional)_ The minimum value, defaults to `0.0`. Must be less than `max`. |
| `max` | _(optional)_ The maximum value, defaults to `1.0` (if no args provided). |

### Examples

```console
$ gomplate -i '{{ random.Float }}'
0.2029946480303966
```
```console
$ gomplate -i '{{ random.Float 100 }}'  
71.28595374161743
```
```console
$ gomplate -i '{{ random.Float -100 200 }}'
105.59119437834909
```
