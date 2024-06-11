---
title: conversion functions
menu:
  main:
    parent: functions
---

These are a collection of functions that mostly help converting from one type
to another - generally from a `string` to something else, and vice-versa.

## `conv.Bool`

**Alias:** `bool`

**Note:** See also [`conv.ToBool`](#conv-tobool) for a more flexible variant.

Converts a true-ish string to a boolean. Can be used to simplify conditional statements based on environment variables or other text input.

_Added in gomplate [v0.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v0.2.0)_
### Usage

```
conv.Bool in
```
```
in | conv.Bool
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the input string |

### Examples

_`input.tmpl`:_
```
{{if bool (getenv "FOO")}}foo{{else}}bar{{end}}
```

```console
$ gomplate < input.tmpl
bar
$ FOO=true gomplate < input.tmpl
foo
```

## `conv.Default`

**Alias:** `default`

Provides a default value given an empty input. Empty inputs are `0` for numeric
types, `""` for strings, `false` for booleans, empty arrays/maps, and `nil`.

Note that this will not provide a default for the case where the input is undefined
(i.e. referencing things like `.foo` where there is no `foo` field of `.`), but
[`conv.Has`](#conv-has) can be used for that.

_Added in gomplate [v2.5.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.5.0)_
### Usage

```
conv.Default default in
```
```
in | conv.Default default
```

### Arguments

| name | description |
|------|-------------|
| `default` | _(required)_ the default value |
| `in` | _(required)_ the input |

### Examples

```console
$ gomplate -i '{{ "" | default "foo" }} {{ "bar" | default "baz" }}'
foo bar
```

## `conv.Dict` _(deprecated)_
**Deprecation Notice:** Renamed to [`coll.Dict`](#coll-dict)

**Alias:** `dict`

Dict is a convenience function that creates a map with string keys.
Provide arguments as key/value pairs. If an odd number of arguments
is provided, the last is used as the key, and an empty string is
set as the value.

All keys are converted to strings.

This function is equivalent to [Sprig's `dict`](http://masterminds.github.io/sprig/dicts.html#dict)
function, as used in [Helm templates](https://docs.helm.sh/chart_template_guide#template-functions-and-pipelines).

For creating more complex maps, see [`data.JSON`](../data/#data-json) or [`data.YAML`](../data/#data-yaml).

For creating arrays, see [`coll.Slice`](#coll-slice).

_Added in gomplate [v3.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.0.0)_
### Usage

```
conv.Dict in...
```

### Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ The key/value pairs |

### Examples

```console
$ gomplate -i '{{ conv.Dict "name" "Frank" "age" 42 | data.ToYAML }}'
age: 42
name: Frank
$ gomplate -i '{{ dict 1 2 3 | toJSON }}'
{"1":2,"3":""}
```
```console
$ cat <<EOF| gomplate
{{ define "T1" }}Hello {{ .thing }}!{{ end -}}
{{ template "T1" (dict "thing" "world")}}
{{ template "T1" (dict "thing" "everybody")}}
EOF
Hello world!
Hello everybody!
```

## `conv.Slice` _(deprecated)_
**Deprecation Notice:** Renamed to [`coll.Slice`](#coll-slice)

**Alias:** `slice`

Creates a slice (like an array or list). Useful when needing to `range` over a bunch of variables.

_Added in gomplate [v0.3.0](https://github.com/hairyhenderson/gomplate/releases/tag/v0.3.0)_
### Usage

```
conv.Slice in...
```

### Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the elements of the slice |

### Examples

```console
$ gomplate -i '{{ range coll.Slice "Bart" "Lisa" "Maggie" }}Hello, {{ . }}{{ end }}'
Hello, Bart
Hello, Lisa
Hello, Maggie
```

## `conv.Has` _(deprecated)_
**Deprecation Notice:** Renamed to [`coll.Has`](#coll-has)

**Alias:** `has`

Reports whether a given object has a property with the given key, or whether a given array/slice contains the given value. Can be used with `if` to prevent the template from trying to access a non-existent property in an object.

_Added in gomplate [v1.5.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.5.0)_
### Usage

```
conv.Has in item
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The object or list to search |
| `item` | _(required)_ The item to search for |

### Examples

```console
$ gomplate -i '{{ $l := coll.Slice "foo" "bar" "baz" }}there is {{ if has $l "bar" }}a{{else}}no{{end}} bar'
there is a bar
```
```console
$ export DATA='{"foo": "bar"}'
$ gomplate -i '{{ $o := data.JSON (getenv "DATA") -}}
{{ if (has $o "foo") }}{{ $o.foo }}{{ else }}THERE IS NO FOO{{ end }}'
bar
```
```console
$ export DATA='{"baz": "qux"}'
$ gomplate -i '{{ $o := data.JSON (getenv "DATA") -}}
{{ if (has $o "foo") }}{{ $o.foo }}{{ else }}THERE IS NO FOO{{ end }}'
THERE IS NO FOO
```

## `conv.Join`

**Alias:** `join`

Concatenates the elements of an array to create a string. The separator string `sep` is placed between elements in the resulting string.

_Added in gomplate [v0.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v0.4.0)_
### Usage

```
conv.Join in sep
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the array or slice |
| `sep` | _(required)_ the separator |

### Examples

```console
$ gomplate -i '{{ $a := coll.Slice 1 2 3 }}{{ join $a "-" }}'
1-2-3
```

## `conv.URL`

**Alias:** `urlParse`

Parses a string as a URL for later use. Equivalent to [url.Parse](https://golang.org/pkg/net/url/#Parse)

Any of `url.URL`'s methods can be called on the result.

_Added in gomplate [v2.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.0.0)_
### Usage

```
conv.URL in
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the URL string to parse |

### Examples

_`input.tmpl`:_
```
{{ $u := conv.URL "https://example.com:443/foo/bar" }}
The scheme is {{ $u.Scheme }}
The host is {{ $u.Host }}
The path is {{ $u.Path }}
```

```console
$ gomplate < input.tmpl
The scheme is https
The host is example.com:443
The path is /foo/bar
```
_Call `Redacted` to hide the password in the output:_
```
$ gomplate -i '{{ (conv.URL "https://user:supersecret@example.com").Redacted }}'
https://user:xxxxx@example.com
```

## `conv.ParseInt`

_**Note:**_ See [`conv.ToInt64`](#conv-toint64) instead for a simpler and more flexible variant of this function.

Parses a string as an int64. Equivalent to [strconv.ParseInt](https://golang.org/pkg/strconv/#ParseInt)

_Added in gomplate [v1.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.4.0)_
### Usage

```
conv.ParseInt
```


### Examples

_`input.tmpl`:_
```
{{ $val := conv.ParseInt (getenv "HEXVAL") 16 32 }}
The value in decimal is {{ $val }}
```

```console
$ HEXVAL=7C0 gomplate < input.tmpl

The value in decimal is 1984
```

## `conv.ParseFloat`

_**Note:**_ See [`conv.ToFloat`](#conv-tofloat) instead for a simpler and more flexible variant of this function.

Parses a string as an float64 for later use. Equivalent to [strconv.ParseFloat](https://golang.org/pkg/strconv/#ParseFloat)

_Added in gomplate [v1.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.4.0)_
### Usage

```
conv.ParseFloat
```


### Examples

_`input.tmpl`:_
```
{{ $pi := conv.ParseFloat (getenv "PI") 64 }}
{{- if (gt $pi 3.0) -}}
pi is greater than 3
{{- end }}
```

```console
$ PI=3.14159265359 gomplate < input.tmpl
pi is greater than 3
```

## `conv.ParseUint`

Parses a string as an uint64 for later use. Equivalent to [strconv.ParseUint](https://golang.org/pkg/strconv/#ParseUint)

_Added in gomplate [v1.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.4.0)_
### Usage

```
conv.ParseUint
```


### Examples

_`input.tmpl`:_
```
{{ conv.ParseInt (getenv "BIG") 16 64 }} is max int64
{{ conv.ParseUint (getenv "BIG") 16 64 }} is max uint64
```

```console
$ BIG=FFFFFFFFFFFFFFFF gomplate < input.tmpl
9223372036854775807 is max int64
18446744073709551615 is max uint64
```

## `conv.Atoi`

_**Note:**_ See [`conv.ToInt`](#conv-toint) and [`conv.ToInt64`](#conv-toint64) instead for simpler and more flexible variants of this function.

Parses a string as an int for later use. Equivalent to [strconv.Atoi](https://golang.org/pkg/strconv/#Atoi)

_Added in gomplate [v1.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v1.4.0)_
### Usage

```
conv.Atoi
```


### Examples

_`input.tmpl`:_
```
{{ $number := conv.Atoi (getenv "NUMBER") }}
{{- if (gt $number 5) -}}
The number is greater than 5
{{- else -}}
The number is less than 5
{{- end }}
```

```console
$ NUMBER=21 gomplate < input.tmpl
The number is greater than 5
```

## `conv.ToBool`

Converts the input to a boolean value.
Possible `true` values are: `1` or the strings `"t"`, `"true"`, or `"yes"`
(any capitalizations). All other values are considered `false`.

_Added in gomplate [v2.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.7.0)_
### Usage

```
conv.ToBool input
```
```
input | conv.ToBool
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ The input to convert |

### Examples

```console
$ gomplate -i '{{ conv.ToBool "yes" }} {{ conv.ToBool true }} {{ conv.ToBool "0x01" }}'
true true true
$ gomplate -i '{{ conv.ToBool false }} {{ conv.ToBool "blah" }} {{ conv.ToBool 0 }}'
false false false
```

## `conv.ToBools`

Converts a list of inputs to an array of boolean values.
Possible `true` values are: `1` or the strings `"t"`, `"true"`, or `"yes"`
(any capitalizations). All other values are considered `false`.

_Added in gomplate [v2.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.7.0)_
### Usage

```
conv.ToBools input
```
```
input | conv.ToBools
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ The input array to convert |

### Examples

```console
$ gomplate -i '{{ conv.ToBools "yes" true "0x01" }}'
[true true true]
$ gomplate -i '{{ conv.ToBools false "blah" 0 }}'
[false false false]
```

## `conv.ToInt64`

Converts the input to an `int64` (64-bit signed integer).

This function attempts to convert most types of input (strings, numbers,
and booleans), but behaviour when the input can not be converted is
undefined and subject to change.

Unconvertable inputs will result in errors.

Floating-point numbers (with decimal points) are truncated.

_Added in gomplate [v2.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.2.0)_
### Usage

```
conv.ToInt64 in
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the value to convert |

### Examples

```console
$ gomplate -i '{{conv.ToInt64 "9223372036854775807"}}'
9223372036854775807
```
```console
$ gomplate -i '{{conv.ToInt64 "0x42"}}'
66
```
```console
$ gomplate -i '{{conv.ToInt64 true }}'
1
```

## `conv.ToInt`

Converts the input to an `int` (signed integer, 32- or 64-bit depending
on platform). This is similar to [`conv.ToInt64`](#conv-toint64) on 64-bit
platforms, but is useful when input to another function must be provided
as an `int`.

Unconvertable inputs will result in errors.

On 32-bit systems, given a number that is too large to fit in an `int`,
the result is `-1`. This is done to protect against
[CWE-190](https://cwe.mitre.org/data/definitions/190.html) and
[CWE-681](https://cwe.mitre.org/data/definitions/681.html).

See also [`conv.ToInt64`](#conv-toint64).

_Added in gomplate [v2.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.2.0)_
### Usage

```
conv.ToInt in
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the value to convert |

### Examples

```console
$ gomplate -i '{{conv.ToInt "9223372036854775807"}}'
9223372036854775807
```
```console
$ gomplate -i '{{conv.ToInt "0x42"}}'
66
```
```console
$ gomplate -i '{{conv.ToInt true }}'
1
```

## `conv.ToInt64s`

Converts the inputs to an array of `int64`s.

Unconvertable inputs will result in errors.

This delegates to [`conv.ToInt64`](#conv-toint64) for each input argument.

_Added in gomplate [v2.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.2.0)_
### Usage

```
conv.ToInt64s in...
```

### Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the inputs to be converted |

### Examples

```console
gomplate -i '{{ conv.ToInt64s true 0x42 "123,456.99" "1.2345e+3"}}'
[1 66 123456 1234]
```

## `conv.ToInts`

Converts the inputs to an array of `int`s.

Unconvertable inputs will result in errors.

This delegates to [`conv.ToInt`](#conv-toint) for each input argument.

_Added in gomplate [v2.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.2.0)_
### Usage

```
conv.ToInts in...
```

### Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the inputs to be converted |

### Examples

```console
gomplate -i '{{ conv.ToInts true 0x42 "123,456.99" "1.2345e+3"}}'
[1 66 123456 1234]
```

## `conv.ToFloat64`

Converts the input to a `float64`.

This function attempts to convert most types of input (strings, numbers,
and booleans), but behaviour when the input can not be converted is
undefined and subject to change.

Unconvertable inputs will result in errors.

_Added in gomplate [v2.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.2.0)_
### Usage

```
conv.ToFloat64 in
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the value to convert |

### Examples

```console
$ gomplate -i '{{ conv.ToFloat64 "8.233e-1"}}'
0.8233
$ gomplate -i '{{ conv.ToFloat64 "9,000.09"}}'
9000.09
```

## `conv.ToFloat64s`

Converts the inputs to an array of `float64`s.

Unconvertable inputs will result in errors.

This delegates to [`conv.ToFloat64`](#conv-tofloat64) for each input argument.

_Added in gomplate [v2.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.2.0)_
### Usage

```
conv.ToFloat64s in...
```

### Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the inputs to be converted |

### Examples

```console
$ gomplate -i '{{ conv.ToFloat64s true 0x42 "123,456.99" "1.2345e+3"}}'
[1 66 123456.99 1234.5]
```

## `conv.ToString`

Converts the input (of any type) to a `string`.

The input will always be represented in _some_ way.

_Added in gomplate [v2.5.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.5.0)_
### Usage

```
conv.ToString in
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the value to convert |

### Examples

```console
$ gomplate -i '{{ conv.ToString 0xFF }}'
255
$ gomplate -i '{{ dict "foo" "bar" | conv.ToString}}'
map[foo:bar]
$ gomplate -i '{{ conv.ToString nil }}'
nil
```

## `conv.ToStrings`

Converts the inputs (of any type) to an array of `string`s

This delegates to [`conv.ToString`](#conv-tostring) for each input argument.

_Added in gomplate [v2.5.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.5.0)_
### Usage

```
conv.ToStrings in...
```

### Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the inputs to be converted |

### Examples

```console
$ gomplate -i '{{ conv.ToStrings nil 42 true 0xF (coll.Slice 1 2 3) }}'
[nil 42 true 15 [1 2 3]]
```
