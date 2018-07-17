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

#### Example

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

#### Example

```console
$ gomplate -i '{{ "" | default "foo" }} {{ "bar" | default "baz" }}'
foo bar
```

## `conv.Slice`

**Alias:** `slice`

Creates a slice. Useful when needing to `range` over a bunch of variables.

#### Example

_`input.tmpl`:_
```
{{range slice "Bart" "Lisa" "Maggie"}}
Hello, {{.}}
{{- end}}
```

```console
$ gomplate < input.tmpl
Hello, Bart
Hello, Lisa
Hello, Maggie
```

## `conv.Has`

**Alias:** `has`

Has reports whether or not a given object has a property with the given key. Can be used with `if` to prevent the template from trying to access a non-existent property in an object.

#### Example

_Let's say we're using a Vault datasource..._

_`input.tmpl`:_
```
{{ $secret := datasource "vault" "mysecret" -}}
The secret is '
{{- if (has $secret "value") }}
{{- $secret.value }}
{{- else }}
{{- $secret | toYAML }}
{{- end }}'
```

If the `secret/foo/mysecret` secret in Vault has a property named `value` set to `supersecret`:

```console
$ gomplate -d vault:///secret/foo < input.tmpl
The secret is 'supersecret'
```

On the other hand, if there is no `value` property:

```console
$ gomplate -d vault:///secret/foo < input.tmpl
The secret is 'foo: bar'
```

## `conv.Join`

**Alias:** `join`

Concatenates the elements of an array to create a string. The separator string sep is placed between elements in the resulting string.

#### Example

_`input.tmpl`_
```
{{ $a := `[1, 2, 3]` | jsonArray }}
{{ join $a "-" }}
```

```console
$ gomplate -f input.tmpl
1-2-3
```


## `conv.URL`

**Alias:** `urlParse`

Parses a string as a URL for later use. Equivalent to [url.Parse](https://golang.org/pkg/net/url/#Parse)

#### Example

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

## `conv.ParseInt`

Parses a string as an int64 for later use. Equivalent to [strconv.ParseInt](https://golang.org/pkg/strconv/#ParseInt)

#### Example

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

Parses a string as an float64 for later use. Equivalent to [strconv.ParseFloat](https://golang.org/pkg/strconv/#ParseFloat)

#### Example

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

Parses a string as an int for later use. Equivalent to [strconv.Atoi](https://golang.org/pkg/strconv/#Atoi)

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

### Usage
```go
conv.ToBool input 
```
```go
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

### Usage
```go
conv.ToBools input 
```

```go
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

Converts the input to an `int64`.

```console
$ gomplate -i '{{conv.ToInt64 "9223372036854775807"}}'
9223372036854775807
```

## `conv.ToInt`

Converts the input to an `int`. This is similar to `conv.Atoi`, but handles booleans and numbers as well as strings.

```console
$ gomplate -i '{{conv.ToInt (gt 1 2)}}'
0
```

## `conv.ToInt64s`

Converts the inputs to an array of `int64`s

## `conv.ToInts`

Converts the inputs to an array of `int`s

## `conv.ToFloat64`

Converts the input to a `float64`

## `conv.ToFloat64s`

Converts the inputs to an array of `float64`s

## `conv.ToString`

Converts the input (of any type) to a `string`

## `conv.ToStrings`

Converts the inputs (of any type) to an array of `string`s
