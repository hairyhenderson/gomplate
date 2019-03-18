---
title: uuid functions
menu:
  main:
    parent: functions
---

Functions for generating, parsing, and manipulating UUIDs.

A UUID is a 128 bit (16 byte) _Universal Unique IDentifier_ as defined
in [RFC 4122][]. Only RFC 4112-variant UUIDs can be generated, but all variants
(even invalid ones) can be parsed and manipulated. Also, gomplate only supports
generating version 1 and 4 UUIDs (with 4 being the most commonly-used variety
these days). Versions 2, 3, and 5 are able to be supported: [log an issue][] if
this is required for your use-case.

[RFC 4122]: https://en.wikipedia.org/wiki/Universally_unique_identifier
[log an issue]: https://github.com/hairyhenderson/gomplate/issues/new

## `uuid.V1`

Create a version 1 UUID (based on the current MAC address and the current date/time).

Use [`uuid.V4`](#uuid-v4) instead in most cases.

### Usage

```go
uuid.V1
```


### Examples

```console
$ gomplate -i '{{ uuid.V1 }}'
4d757e54-446d-11e9-a8fa-72000877c7b0
```

## `uuid.V4`

Create a version 4 UUID (randomly generated).

This function consumes entropy.

### Usage

```go
uuid.V4
```


### Examples

```console
$ gomplate -i '{{ uuid.V4 }}'
40b3c2d2-e491-4b19-94cd-461e6fa35a60
```

## `uuid.Nil`

Returns the _nil_ UUID, that is, `00000000-0000-0000-0000-000000000000`,
mostly for testing scenarios.

### Usage

```go
uuid.Nil
```


### Examples

```console
$ gomplate -i '{{ uuid.Nil }}'
00000000-0000-0000-0000-000000000000
```

## `uuid.IsValid`

Checks that the given UUID is in the correct format. It does not validate
whether the version or variant are correct.

### Usage

```go
uuid.IsValid uuid
```
```go
uuid | uuid.IsValid
```

### Arguments

| name | description |
|------|-------------|
| `uuid` | _(required)_ The uuid to check |

### Examples

```console
$ gomplate -i '{{ if uuid.IsValid "totally invalid" }}valid{{ else }}invalid{{ end }}'
invalid
```
```console
$ gomplate -i '{{ uuid.IsValid "urn:uuid:12345678-90ab-cdef-fedc-ba9876543210" }}'
true
```

## `uuid.Parse`

Parse a UUID for further manipulation or inspection.

This function returns a `UUID` struct, as defined in the [github.com/google/uuid](https://godoc.org/github.com/google/uuid#UUID) package. See the docs for examples of functions or fields you can call.

Both the standard UUID forms of `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` and
`urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` are decoded as well as the
Microsoft encoding `{xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx}` and the raw hex
encoding (`xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`).

### Usage

```go
uuid.Parse uuid
```
```go
uuid | uuid.Parse
```

### Arguments

| name | description |
|------|-------------|
| `uuid` | _(required)_ The uuid to parse |

### Examples

```console
$ gomplate -i '{{ $u := uuid.Parse uuid.V4 }}{{ $u.Version }}, {{ $u.Variant}}'
VERSION_4, RFC4122
```
```console
$ gomplate -i '{{ (uuid.Parse "000001f5-4470-21e9-9b00-72000877c7b0").Domain }}'
Person
```
