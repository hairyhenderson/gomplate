---
title: math functions
menu:
  main:
    parent: functions
---

A set of basic math functions to be able to perform simple arithmetic operations with `gomplate`.

### Supported input

_**Note:** currently, `gomplate` supports only integer arithmetic. All functions
return 64-bit integers (`int64` type). Floating point support will be added in
later releases._

In general, any input will be converted to the correct input type by the various
functions in this package. For integer-based functions, floating-point inputs will
be truncated (not rounded).

In addition to regular base-10 numbers, integers can be
[specified](https://golang.org/ref/spec#Integer_literals) as octal (prefix with
`0`) or hexadecimal (prefix with `0x`).

Decimal/floating-point numbers can be [specified](https://golang.org/ref/spec#Floating-point_literals)
with optional exponents.

Some examples demonstrating this:

```console
$ NUM=50 gomplate -i '{{ div (getenv "NUM") 10 }}'
5
$ gomplate -i '{{ add "0x2" "02" "2.0" "2e0" }}'
8
$ gomplate -i '{{ add 2.5 2.5 }}' # decimals are truncated!
4
```

## `math.Add`

**Alias:** `add`

Adds all given operators.

### Usage
```go
math.Add n...
```
```go
x | math.Add.Add n...
```

### Example

```console
$ gomplate -i '{{ math.Add 1 2 3 4 }}
10
```

## `math.Sub`

**Alias:** `sub`

Subtract the second from the first of the given operators.

### Usage
```go
math.Sub a b
```
```go
b | math.Sub a
```

### Example

```console
$ gomplate -i '{{ math.Sub 3 1 }}'
2
```

## `math.Mul`

**Alias:** `mul`

Multiply all given operators together.

### Usage
```go
math.Mul n...
```
```go
x | math.Mul n...
```

### Example

```console
$ gomplate -i '{{ math.Mul 8 8 2 }}'
128
```

## `math.Div`

**Alias:** `div`

Divide the first number by the second. Division by zero is disallowed.

### Usage
```go
math.Div a b
```
```go
b | math.Div a
```

### Example

```console
$ gomplate -i '{{ math.Div 8 2 }}'
4
```

## `math.Rem`

**Alias:** `rem`

Return the remainder from an integer division operation.

### Usage
```go
math.Rem a b
```
```go
b | math.Rem b
```

### Example

```console
$ gomplate -i '{{ math.Rem 5 3 }}'
2
$ gomplate -i '{{ math.Rem -5 3 }}'
-2
```

## `math.Pow`

**Alias:** `pow`

Calculate an exponent - _b<sup>n</sup>_. This wraps Go's [`math.Pow`](https://golang.org/pkg/math/#Pow).

### Usage
```go
math.Pow b n
```
```go
n | math.Pow b
```

### Example

```console
$ gomplate -i '{{ math.Pow 10 2 }}'
100
$ gomplate -i '{{ math.Pow 2 32 }}'
4294967296
```

## `math.Seq`

**Alias:** `seq`

Return a sequence from `start` to `end`, in steps of `step`. Can handle counting
down as well as up, including with negative numbers.

Note that the sequence _may_ not end at `end`, if `end` is not divisible by `step`.

### Usage
```go
math.Seq [start] end [step]
```

### Arguments

| name   | description |
|--------|-------|
| `start` | _(optional)_ The first number in the sequence (defaults to `1`) |
| `end` | _(required)_ The last number in the sequence |
| `step` | _(optional)_ The amount to increment between each number (defaults to `1`) |

### Examples

```console
$ gomplate -i '{{ range (math.Seq 5) }}{{.}} {{end}}'
1 2 3 4 5 
```

```console
$ gomplate -i '{{ conv.Join (math.Seq 10 -3 2) ", " }}'
10, 8, 6, 4, 2, 0, -2
```