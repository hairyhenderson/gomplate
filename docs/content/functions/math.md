---
title: math functions
menu:
  main:
    parent: functions
---

A set of basic math functions to be able to perform simple arithmetic operations with `gomplate`.

### Supported input

In general, any input will be converted to the correct input type by the various functions in this package, and an appropriately-typed value will be returned type. Special cases are documented.

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
$ gomplate -i '{{ add 2.5 2.5 }}'
5.0
```

## `math.Abs`

Returns the absolute value of a given number. When the input is an integer, , the result will be an `int64`, otherwise it will be an `float64`.

### Usage
```go
math.Abs num
```

### Arguments

| name   | description |
|--------|-------|
| `num`  | _(required)_ The input number |

### Examples

```console
$ gomplate -i '{{ math.Abs -3.5 }} {{ math.Abs 3.5 }} {{ math.Abs -42 }}'
3.5 3.5 42
```

## `math.Add`

**Alias:** `add`

Adds all given operators. When one of the inputs is a floating-point number, the result will be a `float64`, otherwise it will be an `int64`.

### Usage
```go
math.Add n...
```
```go
x | math.Add.Add n...
```

### Example

```console
$ gomplate -i '{{ math.Add 1 2 3 4 }} {{ math.Add 1.5 2 3 }}'
10 6.5
```

## `math.Ceil`

Returns the least integer value greater than or equal to a given floating-point number. This wraps Go's [`math.Ceil`](https://golang.org/pkg/math/#Ceil).

**Note:** the return value of this function is a `float64` so that the special-cases `NaN` and `Inf` can be returned appropriately.

### Usage
```go
math.Ceil num
```

### Arguments

| name   | description |
|--------|-------|
| `num` | _(required)_ The input number. Will be converted to a `float64`, or `0` if not convertable |

### Examples

```console
$ gomplate -i '{{ range (slice 5.1 42 "3.14" "0xFF" "NaN" "Inf" "-0") }}ceil {{ printf "%#v" . }} = {{ math.Ceil . }}{{"\n"}}{{ end }}' 
ceil 5.1 = 6
ceil 42 = 42
ceil "3.14" = 4
ceil "0xFF" = 255
ceil "NaN" = NaN
ceil "Inf" = +Inf
ceil "-0" = 0
```

## `math.Div`

**Alias:** `div`

Divide the first number by the second. Division by zero is disallowed. The result will be a `float64`.

### Usage
```go
math.Div a b
```
```go
b | math.Div a
```

### Example

```console
$ gomplate -i '{{ math.Div 8 2 }} {{ math.Div 3 2 }}'
4 1.5
```

## `math.Floor`

Returns the greatest integer value less than or equal to a given floating-point number. This wraps Go's [`math.Floor`](https://golang.org/pkg/math/#Floor).

**Note:** the return value of this function is a `float64` so that the special-cases `NaN` and `Inf` can be returned appropriately.

### Usage
```go
math.Floor num
```

### Arguments

| name   | description |
|--------|-------|
| `num` | _(required)_ The input number. Will be converted to a `float64`, or `0` if not convertable |

### Examples

```console
$ gomplate -i '{{ range (slice 5.1 42 "3.14" "0xFF" "NaN" "Inf" "-0") }}floor {{ printf "%#v" . }} = {{ math.Floor . }}{{"\n"}}{{ end }}'
floor 5.1 = 4
floor 42 = 42
floor "3.14" = 3
floor "0xFF" = 255
floor "NaN" = NaN
floor "Inf" = +Inf
floor "-0" = 0
```

## `math.IsFloat`

Returns whether or not the given number can be interpreted as a floating-point literal, as defined by the [Go language reference](https://golang.org/ref/spec#Floating-point_literals).

**Note:** If a decimal point is part of the input number, it will be considered a floating-point number, even if the decimal is `0`.

### Usage
```go
math.IsFloat num
```

### Arguments

| name   | description |
|--------|-------|
| `num` | _(required)_ The value to test |

### Examples

```console
$ gomplate -i '{{ range (slice 1.0 "-1.0" 5.1 42 "3.14" "foo" "0xFF" "NaN" "Inf" "-0") }}{{ if (math.IsFloat .) }}{{.}} is a float{{"\n"}}{{ end }}{{end}}'
1 is a float
-1.0 is a float
5.1 is a float
3.14 is a float
NaN is a float
Inf is a float
```

## `math.IsInt`

Returns whether or not the given number is an integer.
Returns whether or not the given number can be interpreted as a floating-point literal, as defined by the [Go language reference](https://golang.org/ref/spec#Integer_literals).

### Usage
```go
math.IsInt num
```

### Arguments

| name   | description |
|--------|-------|
| `num` | _(required)_ The value to test |

### Examples

```console
$ gomplate -i '{{ range (slice 1.0 "-1.0" 5.1 42 "3.14" "foo" "0xFF" "NaN" "Inf" "-0") }}{{ if (math.IsInt .) }}{{.}} is an integer{{"\n"}}{{ end }}{{end}}'
42 is an integer
0xFF is an integer
-0 is an integer
```

## `math.IsNum`

Returns whether the given input is a number. Useful for `if` conditions.

### Usage
```go
math.IsNum in
```

### Arguments

| name   | description |
|--------|-------|
| `in` | _(required)_ The value to test |

### Examples

```console
$ gomplate -i '{{ math.IsNum "foo" }} {{ math.IsNum 0xDeadBeef }}'
false true
```

## `math.Max`

Returns the largest number provided. If any values are floating-point numbers, a `float64` is returned, otherwise an `int64` is returned. The same special-cases as Go's [`math.Max`](https://golang.org/pkg/math/#Max) are followed.

### Usage
```go
math.Max nums...
```

### Arguments

| name   | description |
|--------|-------|
| `nums` | _(required)_ One or more numbers to compare |

### Examples

```console
$ gomplate -i '{{ math.Max 0 8.0 4.5 "-1.5e-11" }}'
8
```

## `math.Min`

Returns the smallest number provided. If any values are floating-point numbers, a `float64` is returned, otherwise an `int64` is returned. The same special-cases as Go's [`math.Min`](https://golang.org/pkg/math/#Min) are followed.

### Usage
```go
math.Min nums...
```

### Arguments

| name   | description |
|--------|-------|
| `nums` | _(required)_ One or more numbers to compare |

### Examples

```console
$ gomplate -i '{{ math.Min 0 8 4.5 "-1.5e-11" }}'
-1.5e-11
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

## `math.Pow`

**Alias:** `pow`

Calculate an exponent - _b<sup>n</sup>_. This wraps Go's [`math.Pow`](https://golang.org/pkg/math/#Pow). If any values are floating-point numbers, a `float64` is returned, otherwise an `int64` is returned.

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
$ gomplate -i '{{ math.Pow 1.5 2 }}'
2.2
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

## `math.Round`

Returns the nearest integer, rounding half away from zero.

**Note:** the return value of this function is a `float64` so that the special-cases `NaN` and `Inf` can be returned appropriately.

### Usage
```go
math.Round num
```

### Arguments

| name   | description |
|--------|-------|
| `num` | _(required)_ The input number. Will be converted to a `float64`, or `0` if not convertable |

### Examples

```console
$ gomplate -i '{{ range (slice -6.5 5.1 42.9 "3.5" 6.5) }}round {{ printf "%#v" . }} = {{ math.Round . }}{{"\n"}}{{ end }}'
round -6.5 = -7
round 5.1 = 5
round 42.9 = 43
round "3.5" = 4
round 6.5 = 7
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

## `math.Sub`

**Alias:** `sub`

Subtract the second from the first of the given operators.  When one of the inputs is a floating-point number, the result will be a `float64`, otherwise it will be an `int64`.

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
