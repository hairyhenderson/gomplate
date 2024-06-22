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

_Added in gomplate [v3.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.2.0)_
### Usage

```
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

## `coll.Slice` _(deprecated)_
**Deprecation Notice:** The `slice` alias is deprecated, use the full name `coll.Slice` instead.

**Alias:** `slice`

Creates a slice (like an array or list). Useful when needing to `range` over a bunch of variables.

_Added in gomplate [v3.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.2.0)_
### Usage

```
coll.Slice in...
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

## `coll.GoSlice`

This exposes the `slice` function from Go's [`text/template`](https://golang.org/pkg/text/template/#hdr-Functions)
package. Note that using `slice` will use the `coll.Slice` function instead,
which may not be desired.
For some background on this, see [this issue](https://github.com/hairyhenderson/gomplate/issues/1461).

Here is the upstream documentation:

```
slice returns the result of slicing its first argument by the
remaining arguments. Thus "slice x 1 2" is, in Go syntax, x[1:2],
while "slice x" is x[:], "slice x 1" is x[1:], and "slice x 1 2 3"
is x[1:2:3]. The first argument must be a string, slice, or array.
```

See the [Go language spec](https://go.dev/ref/spec#Slice_expressions) for
more details.

_Added in gomplate [v4.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v4.0.0)_
### Usage

```
coll.GoSlice item [indexes...]
```

### Arguments

| name | description |
|------|-------------|
| `item` | _(required)_ the string, slice, or array to slice |
| `indexes...` | _(optional)_ the indexes to slice the item by (0 to 3 arguments) |

### Examples

```console
$ gomplate -i '{{ coll.GoSlice "hello world" 3 8 }}'
lo wo
```

## `coll.Has`

**Alias:** `has`

Reports whether a given object has a property with the given key, or whether a given array/slice contains the given value. Can be used with `if` to prevent the template from trying to access a non-existent property in an object.

_Added in gomplate [v3.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.2.0)_
### Usage

```
coll.Has in item
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

## `coll.Index`

Returns the result of indexing the given map, slice, or array by the given
key or index. This is similar to the built-in `index` function, but the
arguments are ordered differently for pipeline compatibility. Also this
function is more strict, and will return an error when trying to access a
non-existent map key.

Multiple indexes may be given, for nested indexing.

_Added in gomplate [v4.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v4.0.0)_
### Usage

```
coll.Index indexes... in
```
```
in | coll.Index indexes...
```

### Arguments

| name | description |
|------|-------------|
| `indexes...` | _(required)_ The key or index |
| `in` | _(required)_ The map, slice, or array to index |

### Examples

```console
$ gomplate -i '{{ coll.Index "foo" (dict "foo" "bar") }}'
bar
```
```console
$ gomplate -i '{{ $m := json `{"foo": {"bar": "baz"}}` -}}
  {{ coll.Index "foo" "bar" $m }}'
baz
```
```console
$ gomplate -i '{{ coll.Slice "foo" "bar" "baz" | coll.Index 1 }}'
bar
```

## `coll.JSONPath`

**Alias:** `jsonpath`

Extracts portions of an input object or list using a [JSONPath][] expression.

Any object or list may be used as input. The output depends somewhat on the expression; if multiple items are matched, an array is returned.

JSONPath expressions can be validated at https://jsonpath.com

[JSONPath]: https://goessner.net/articles/JsonPath

_Added in gomplate [v3.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.4.0)_
### Usage

```
coll.JSONPath expression in
```
```
in | coll.JSONPath expression
```

### Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The JSONPath expression |
| `in` | _(required)_ The object or list to query |

### Examples

```console
$ gomplate -i '{{ .books | jsonpath `$..works[?( @.edition_count > 400 )].title` }}' -c books=https://openlibrary.org/subjects/fantasy.json
[Alice's Adventures in Wonderland Gulliver's Travels]
```

## `coll.JQ`

**Alias:** `jq`

Filters an input object or list using the [jq](https://stedolan.github.io/jq/) language, as implemented by [gojq](https://github.com/itchyny/gojq).

Any JSON datatype may be used as input (NOTE: strings are not JSON-parsed but passed in as is).
If the expression results in multiple items (no matter if streamed or as an array) they are wrapped in an array.
Otherwise a single item is returned (even if resulting in an array with a single contained element).

JQ filter expressions can be tested at https://jqplay.org/

See also:

- [jq manual](https://stedolan.github.io/jq/manual/)
- [gojq differences to jq](https://github.com/itchyny/gojq#difference-to-jq)

_Added in gomplate [v4.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v4.0.0)_
### Usage

```
coll.JQ expression in
```
```
in | coll.JQ expression
```

### Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The JQ expression |
| `in` | _(required)_ The object or list to query |

### Examples

```console
$ gomplate \
   -i '{{ .books | jq `[.works[]|{"title":.title,"authors":[.authors[].name],"published":.first_publish_year}][0]` }}' \
   -c books=https://openlibrary.org/subjects/fantasy.json
map[authors:[Lewis Carroll] published:1865 title:Alice's Adventures in Wonderland]
```

## `coll.Keys`

**Alias:** `keys`

Return a list of keys in one or more maps.

The keys will be ordered first by map position (if multiple maps are given),
then alphabetically.

See also [`coll.Values`](#coll-values).

_Added in gomplate [v3.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.2.0)_
### Usage

```
coll.Keys in...
```
```
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

_Added in gomplate [v3.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.2.0)_
### Usage

```
coll.Values in...
```
```
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

_Added in gomplate [v3.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.2.0)_
### Usage

```
coll.Append value list...
```
```
list... | coll.Append value
```

### Arguments

| name | description |
|------|-------------|
| `value` | _(required)_ the value to add |
| `list...` | _(required)_ the slice or array to append to |

### Examples

```console
$ gomplate -i '{{ coll.Slice 1 1 2 3 | append 5 }}'
[1 1 2 3 5]
```

## `coll.Prepend`

**Alias:** `prepend`

Prepend a value to the beginning of a list.

_Note that this function does not change the given list; it always produces a new one._

See also [`coll.Append`](#coll-append).

_Added in gomplate [v3.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.2.0)_
### Usage

```
coll.Prepend value list...
```
```
list... | coll.Prepend value
```

### Arguments

| name | description |
|------|-------------|
| `value` | _(required)_ the value to add |
| `list...` | _(required)_ the slice or array to prepend to |

### Examples

```console
$ gomplate -i '{{ coll.Slice 4 3 2 1 | prepend 5 }}'
[5 4 3 2 1]
```

## `coll.Uniq`

**Alias:** `uniq`

Remove any duplicate values from the list, without changing order.

_Note that this function does not change the given list; it always produces a new one._

_Added in gomplate [v3.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.2.0)_
### Usage

```
coll.Uniq list
```
```
list | coll.Uniq
```

### Arguments

| name | description |
|------|-------------|
| `list` | _(required)_ the input list |

### Examples

```console
$ gomplate -i '{{ coll.Slice 1 2 3 2 3 4 1 5 | uniq }}'
[1 2 3 4 5]
```

## `coll.Flatten`

**Alias:** `flatten`

Flatten a nested list. Defaults to completely flattening all nested lists,
but can be limited with `depth`.

_Note that this function does not change the given list; it always produces a new one._

_Added in gomplate [v3.6.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.6.0)_
### Usage

```
coll.Flatten [depth] list
```
```
list | coll.Flatten [depth]
```

### Arguments

| name | description |
|------|-------------|
| `depth` | _(optional)_ maximum depth of nested lists to flatten. Omit or set to `-1` for infinite depth. |
| `list` | _(required)_ the input list |

### Examples

```console
$ gomplate -i '{{ "[[1,2],[],[[3,4],[[[5],6],7]]]" | jsonArray | flatten }}'
[1 2 3 4 5 6 7]
```
```console
$ gomplate -i '{{ coll.Flatten 2 ("[[1,2],[],[[3,4],[[[5],6],7]]]" | jsonArray) }}'
[1 2 3 4 [[5] 6] 7]
```

## `coll.Reverse`

**Alias:** `reverse`

Reverse a list.

_Note that this function does not change the given list; it always produces a new one._

_Added in gomplate [v3.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.2.0)_
### Usage

```
coll.Reverse list
```
```
list | coll.Reverse
```

### Arguments

| name | description |
|------|-------------|
| `list` | _(required)_ the list to reverse |

### Examples

```console
$ gomplate -i '{{ coll.Slice 4 3 2 1 | reverse }}'
[1 2 3 4]
```

## `coll.Sort`

**Alias:** `sort`

Sort a given list. Uses the natural sort order if possible. For inputs
that are not sortable (either because the elements are of different types,
or of an un-sortable type), the input will simply be returned, unmodified.

Maps and structs can be sorted by a named key.

_Note that this function does not modify the input._

_Added in gomplate [v3.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.2.0)_
### Usage

```
coll.Sort [key] list
```
```
list | coll.Sort [key]
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(optional)_ the key to sort by, for lists of maps or structs |
| `list` | _(required)_ the slice or array to sort |

### Examples

```console
$ gomplate -i '{{ coll.Slice "foo" "bar" "baz" | coll.Sort }}'
[bar baz foo]
```
```console
$ gomplate -i '{{ sort (coll.Slice 3 4 1 2 5) }}'
[1 2 3 4 5]
```
```console
$ cat <<EOF > in.json
[{"a": "foo", "b": 1}, {"a": "bar", "b": 8}, {"a": "baz", "b": 3}]
EOF
$ gomplate -d in.json -i '{{ range (include "in" | jsonArray | coll.Sort "b") }}{{ print .a "\n" }}{{ end }}'
foo
baz
bar
```

## `coll.Merge`

**Alias:** `merge`

Merge maps together by overriding src with dst.

In other words, the src map can be configured the "default" map, whereas the dst
map can be configured the "overrides".

Many source maps can be provided. Precedence is in left-to-right order.

_Note that this function does not modify the input._

_Added in gomplate [v3.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.2.0)_
### Usage

```
coll.Merge dst srcs...
```
```
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

## `coll.Pick`

Given a map, returns a new map with any entries that have the given keys.

The keys can either be separate arguments, or a slice (since v4.0.0).

This is the inverse of [`coll.Omit`](#coll-omit).

_Note that this function does not modify the input._

_Added in gomplate [v3.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.7.0)_
### Usage

```
coll.Pick keys... map
```
```
map | coll.Pick keys...
```

### Arguments

| name | description |
|------|-------------|
| `keys...` | _(required)_ the keys (strings) to match |
| `map` | _(required)_ the map to pick from |

### Examples

```console
$ gomplate -i '{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}
{{ coll.Pick "foo" "baz" $data }}'
map[baz:3 foo:1]
```
```console
$ gomplate -i '{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}
{{ $keys := coll.Slice "foo" "baz" }}
{{ coll.Pick $keys $data }}'
map[baz:3 foo:1]
```

## `coll.Omit`

Given a map, returns a new map without any entries that have the given keys.

The keys can either be separate arguments, or a slice (since v4.0.0).

This is the inverse of [`coll.Pick`](#coll-pick).

_Note that this function does not modify the input._

_Added in gomplate [v3.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v3.7.0)_
### Usage

```
coll.Omit keys... map
```
```
map | coll.Omit keys...
```

### Arguments

| name | description |
|------|-------------|
| `keys...` | _(required)_ the keys (strings) to match |
| `map` | _(required)_ the map to omit from |

### Examples

```console
$ gomplate -i '{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}
{{ coll.Omit "foo" "baz" $data }}'
map[bar:2]
```
```console
$ gomplate -i '{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}
{{ $keys := coll.Slice "foo" "baz" }}
{{ coll.Omit $keys $data }}'
map[bar:2]
```

## `coll.Set`

**Alias:** `set`

Sets the given key to the given value in the given map.

The map is modified in place, and the modified map is returned.

_Added in gomplate [v4.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v4.0.0)_
### Usage

```
coll.Set key value map
```
```
map | coll.Set key value
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(required)_ the key (string) to set |
| `value` | _(required)_ the value to set |
| `map` | _(required)_ the map to modify |

### Examples

```console
$ gomplate -i '{{ $data := dict "foo" 1 "bar" 2 }}
{{ coll.Set "baz" 3 $data }}'
map[bar:2 baz:3 foo:1]
```
```console
$ gomplate -i '{{ dict "foo" 1 | coll.Set "bar" 2 }}'
map[bar:2 foo:1]
```

## `coll.Unset`

**Alias:** `unset`

Deletes the element with the specified key in the given map. If there is no such element, the map is returned unchanged.

The map is modified in place, and the modified map is returned.

_Added in gomplate [v4.0.0](https://github.com/hairyhenderson/gomplate/releases/tag/v4.0.0)_
### Usage

```
coll.Unset key map
```
```
map | coll.Unset key
```

### Arguments

| name | description |
|------|-------------|
| `key` | _(required)_ the key (string) to unset |
| `map` | _(required)_ the map to modify |

### Examples

```console
$ gomplate -i '{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}
{{ coll.Unset "bar" $data }}'
map[baz:3 foo:1]
```
```console
$ gomplate -i '{{ dict "foo" 1 "bar" 2 | coll.Unset "bar" }}'
map[foo:1]
```
