---
title: base64 functions
menu:
  main:
    parent: functions
---


## `base64.Encode`

Encode data as a Base64 string. Specifically, this uses the standard Base64 encoding as defined in [RFC4648 &sect;4](https://tools.ietf.org/html/rfc4648#section-4) (and _not_ the URL-safe encoding).

### Usage

```go
base64.Encode input
```
```go
input | base64.Encode
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ The data to encode. Can be a string, a byte array, or a buffer. Other types will be converted to strings first. |

### Examples

```console
$ gomplate -i '{{ base64.Encode "hello world" }}'
aGVsbG8gd29ybGQ=
```
```console
$ gomplate -i '{{ "hello world" | base64.Encode }}'
aGVsbG8gd29ybGQ=
```

## `base64.Decode`

Decode a Base64 string. This supports both standard ([RFC4648 &sect;4](https://tools.ietf.org/html/rfc4648#section-4)) and URL-safe ([RFC4648 &sect;5](https://tools.ietf.org/html/rfc4648#section-5)) encodings.

This function outputs the data as a string, so it may not be appropriate
for decoding binary data. Use [`base64.DecodeBytes`](#base64.DecodeBytes)
for binary data.

### Usage

```go
base64.Decode input
```
```go
input | base64.Decode
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ The base64 string to decode |

### Examples

```console
$ gomplate -i '{{ base64.Decode "aGVsbG8gd29ybGQ=" }}'
hello world
```
```console
$ gomplate -i '{{ "aGVsbG8gd29ybGQ=" | base64.Decode }}'
hello world
```

## `base64.DecodeBytes`

Decode a Base64 string. This supports both standard ([RFC4648 &sect;4](https://tools.ietf.org/html/rfc4648#section-4)) and URL-safe ([RFC4648 &sect;5](https://tools.ietf.org/html/rfc4648#section-5)) encodings.

This function outputs the data as a byte array, so it's most useful for
outputting binary data that will be processed further.
Use [`base64.Decode`](#base64.Decode) to output a plain string.

### Usage

```go
base64.DecodeBytes input
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ The base64 string to decode |

### Examples

```console
$ gomplate -i '{{ base64.DecodeBytes "aGVsbG8gd29ybGQ=" }}'
[104 101 108 108 111 32 119 111 114 108 100]
```
```console
$ gomplate -i '{{ "aGVsbG8gd29ybGQ=" | base64.DecodeBytes | conv.ToString }}'
hello world
```
