---
title: collection functions
menu:
  main:
    parent: functions
---

These functions help manipulate and query collections of data, like lists (slices, or arrays) and maps (dictionaries).

#### Implementation Note
For the functions that return an array, a Go `[]interface{}` is returned, regardless of whether or not the
input was a different type.

## `coll.Dict`

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

### Usage
```go
coll.Dict in... 
```

### Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ The key/value pairs |

### Examples

```console
$ gomplate -i '{{ coll.Dict "name" "Frank" "age" 42 | data.ToYAML }}'
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

## `coll.Slice`

**Alias:** `slice`

Creates a slice (like an array or list). Useful when needing to `range` over a bunch of variables.

### Usage
```go
coll.Slice in... 
```

### Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the elements of the slice |

### Examples

```console
$ gomplate -i '{{ range slice "Bart" "Lisa" "Maggie" }}Hello, {{ . }}{{ end }}'
Hello, Bart
Hello, Lisa
Hello, Maggie
```

## `coll.Has`

**Alias:** `has`

Reports whether a given object has a property with the given key, or whether a given array/slice contains the given value. Can be used with `if` to prevent the template from trying to access a non-existent property in an object.

### Usage
```go
coll.Has in item 
```

### Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The object or list to search |
| `item` | _(required)_ The item to search for |

### Examples

```console
$ gomplate -i '{{ $l := slice "foo" "bar" "baz" }}there is {{ if has $l "bar" }}a{{else}}no{{end}} bar'
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

## `coll.Keys`

**Alias:** `keys`

Return a list of keys in one or more maps.

The keys will be ordered first by map position (if multiple maps are given),
then alphabetically.

See also [`coll.Values`](#coll-values).

### Usage
```go
coll.Keys in... 
```

```go
in... | coll.Keys  
```

### Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the maps |

### Examples

```console
$ gomplate -i '{{ coll.Keys (dict "foo" 1 "bar" 2) }}'
[bar foo]
$ gomplate -i '{{ $map1 := dict "foo" 1 "bar" 2 -}}{{ $map2 := dict "baz" 3 "qux" 4 -}}{{ coll.Keys $map1 $map2 }}'
[bar foo baz qux]
```

## `coll.Values`

**Alias:** `values`

Return a list of values in one or more maps.

The values will be ordered first by map position (if multiple maps are given),
then alphabetically by key.

See also [`coll.Keys`](#coll-keys).

### Usage
```go
coll.Values in... 
```

```go
in... | coll.Values  
```

### Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the maps |

### Examples

```console
$ gomplate -i '{{ coll.Values (dict "foo" 1 "bar" 2) }}'
[2 1]
$ gomplate -i '{{ $map1 := dict "foo" 1 "bar" 2 -}}{{ $map2 := dict "baz" 3 "qux" 4 -}}{{ coll.Values $map1 $map2 }}'
[2 1 3 4]
```

## `coll.Append`

**Alias:** `append`

Append a value to the end of a list.

_Note that this function does not change the given list; it always produces a new one._

See also [`coll.Prepend`](#coll-prepend).

### Usage
```go
coll.Append value list... 
```

```go
list... | coll.Append value  
```

### Arguments

| name | description |
|------|-------------|
| `value` | _(required)_ the value to add |
| `list...` | _(required)_ the slice or array to append to |

### Examples

```console
$ gomplate -i '{{ slice 1 1 2 3 | append 5 }}'
[1 1 2 3 5]
```

## `coll.Prepend`

**Alias:** `prepend`

Prepend a value to the beginning of a list.

_Note that this function does not change the given list; it always produces a new one._

See also [`coll.Append`](#coll-append).

### Usage
```go
coll.Prepend value list... 
```

```go
list... | coll.Prepend value  
```

### Arguments

| name | description |
|------|-------------|
| `value` | _(required)_ the value to add |
| `list...` | _(required)_ the slice or array to prepend to |

### Examples

```console
$ gomplate -i '{{ slice 4 3 2 1 | prepend 5 }}'
[5 4 3 2 1]
```

## `coll.Uniq`

**Alias:** `uniq`

Remove any duplicate values from the list, without changing order.

_Note that this function does not change the given list; it always produces a new one._

See also [`coll.Append`](#coll-append).

### Usage
```go
coll.Uniq list 
```

```go
list | coll.Uniq  
```

### Arguments

| name | description |
|------|-------------|
| `list` | _(required)_ the input list |

### Examples

```console
$ gomplate -i '{{ slice 4 3 2 1 | prepend 5 }}'
[5 4 3 2 1]
```

## `coll.Reverse`

**Alias:** `reverse`

Reverse a list.

_Note that this function does not change the given list; it always produces a new one._

### Usage
```go
coll.Reverse list 
```

```go
list | coll.Reverse  
```

### Arguments

| name | description |
|------|-------------|
| `list` | _(required)_ the list to reverse |

### Examples

```console
$ gomplate -i '{{ slice 4 3 2 1 | reverse }}'
[1 2 3 4]
```

## `coll.Merge`

**Alias:** `merge`

Merge maps together by overriding src with dst.

In other words, the src map can be configured the "default" map, whereas the dst
map can be configured the "overrides".

Many source maps can be provided. Precedence is in left-to-right order.

Note that this function _changes_ the destination map.

### Usage
```go
coll.Merge dst srcs... 
```

```go
srcs... | coll.Merge dst  
```

### Arguments

| name | description |
|------|-------------|
| `dst` | _(required)_ the map to merge _into_ |
| `srcs...` | _(required)_ the map (or maps) to merge _from_ |

### Examples

```console
$ gomplate -i '{{ $default := dict "foo" 1 "bar" 2}}
{{ $config := dict "foo" 8 }}
{{ merge $config $default }}'
map[bar:2 foo:8]
```
```console
$ gomplate -i '{{ $dst := dict "foo" 1 "bar" 2 }}
{{ $src1 := dict "foo" 8 "baz" 4 }}
{{ $src2 := dict "foo" 3 "bar" 5 }}
{{ coll.Merge $dst $src1 $src2 }}'
map[foo:1 bar:5 baz:4]
```
