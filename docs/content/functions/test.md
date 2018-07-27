---
title: test functions
menu:
  main:
    parent: functions
---

The `test` namespace contains some simple functions to help validate
assumptions and can cause template generation to fail in specific cases.

## `test.Assert`

**Alias:** `assert`

Asserts that the given expression or value is `true`. If it is not, causes
template generation to fail immediately with an optional message.

### Usage
```go
test.Assert [message] value 
```

```go
value | test.Assert [message]  
```

### Arguments

| name | description |
|------|-------------|
| `message` | _(optional)_ The optional message to provide in the case of failure |
| `value` | _(required)_ The value to test |

### Examples

```console
$ gomplate -i '{{ assert (eq "foo" "bar") }}'
template: <arg>:1:3: executing "<arg>" at <assert (eq "foo" "ba...>: error calling assert: assertion failed
$ gomplate -i '{{ assert "something horrible happened" false }}'
template: <arg>:1:3: executing "<arg>" at <assert "something ho...>: error calling assert: assertion failed: something horrible happened
```

## `test.Fail`

**Alias:** `fail`

Cause template generation to fail immediately, with an optional message.

### Usage
```go
test.Fail [message] 
```

```go
message | test.Fail  
```

### Arguments

| name | description |
|------|-------------|
| `message` | _(optional)_ The optional message to provide |

### Examples

```console
$ gomplate -i '{{ fail }}'
template: <arg>:1:3: executing "<arg>" at <fail>: error calling fail: template generation failed
$ gomplate -i '{{ test.Fail "something is wrong!" }}'
template: <arg>:1:7: executing "<arg>" at <test.Fail>: error calling Fail: template generation failed: something is wrong!
```
