---
title: path functions
menu:
  main:
    parent: functions
---

The path functions are split into 2 namespaces: 
- `path`, which is useful for manipulating slash-based (`/`) paths, such as in URLs
- `filepath`, which should be used for local filesystem paths, especially when Windows paths may be involved

This page documents the `path` namespace - see also the [`filepath`](../filepath) documentation.

These functions are wrappers for Go's [`path`](https://golang.org/pkg/path/) and [`path/filepath`](https://golang.org/pkg/path/filepath/) packages.

## `path.Base`


Returns the last element of path. Trailing slashes are removed before extracting the last element. If the path is empty, Base returns `.`. If the path consists entirely of slashes, Base returns `/`.

A wrapper for Go's [`path.Base`](https://golang.org/pkg/path/#Base) function.


### Usage
```go
path.Base path 
```

```go
path | path.Base  
```


### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |


### Examples

```console
$ gomplate -i '{{ path.Base "/tmp/foo" }}'
foo
```

## `path.Clean`


Clean returns the shortest path name equivalent to path by purely lexical processing.

A wrapper for Go's [`path.Clean`](https://golang.org/pkg/path/#Clean) function.


### Usage
```go
path.Clean path 
```

```go
path | path.Clean  
```


### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |


### Examples

```console
$ gomplate -i '{{ path.Clean "/tmp//foo/../" }}'
/tmp
```

## `path.Dir`


Returns all but the last element of path, typically the path's directory.

A wrapper for Go's [`path.Dir`](https://golang.org/pkg/path/#Dir) function.


### Usage
```go
path.Dir path 
```

```go
path | path.Dir  
```


### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |


### Examples

```console
$ gomplate -i '{{ path.Dir "/tmp/foo" }}'
/tmp
```

## `path.Ext`


Returns the file name extension used by path.

A wrapper for Go's [`path.Ext`](https://golang.org/pkg/path/#Ext) function.


### Usage
```go
path.Ext path 
```

```go
path | path.Ext  
```


### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |


### Examples

```console
$ gomplate -i '{{ path.Ext "/tmp/foo.csv" }}'
.csv
```

## `path.IsAbs`


Reports whether the path is absolute.

A wrapper for Go's [`path.IsAbs`](https://golang.org/pkg/path/#IsAbs) function.


### Usage
```go
path.IsAbs path 
```

```go
path | path.IsAbs  
```


### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |


### Examples

```console
$ gomplate -i 'the path is {{ if (path.IsAbs "/tmp/foo.csv") }}absolute{{else}}relative{{end}}'
the path is absolute
$ gomplate -i 'the path is {{ if (path.IsAbs "../foo.csv") }}absolute{{else}}relative{{end}}'
the path is relative
```

## `path.Join`


Joins any number of path elements into a single path, adding a separating slash if necessary.

A wrapper for Go's [`path.Join`](https://golang.org/pkg/path/#Join) function.


### Usage
```go
path.Join elem... 
```


### Arguments

| name | description |
|------|-------------|
| `elem...` | _(required)_ The path elements to join (0 or more) |


### Examples

```console
$ gomplate -i '{{ path.Join "/tmp" "foo" "bar" }}'
/tmp/foo/bar
```

## `path.Match`


Reports whether name matches the shell file name pattern.

A wrapper for Go's [`path.Match`](https://golang.org/pkg/path/#Match) function.


### Usage
```go
path.Match pattern path 
```


### Arguments

| name | description |
|------|-------------|
| `pattern` | _(required)_ The pattern to match on |
| `path` | _(required)_ The path to match |


### Examples

```console
$ gomplate -i '{{ path.Match "*.csv" "foo.csv" }}'
true
```

## `path.Split`


Splits path immediately following the final slash, separating it into a directory and file name component.

The function returns an array with two values, the first being the diretory, and the second the file.

A wrapper for Go's [`path.Split`](https://golang.org/pkg/path/#Split) function.


### Usage
```go
path.Split path 
```

```go
path | path.Split  
```


### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |


### Examples

```console
$ gomplate -i '{{ $p := path.Split "/tmp/foo" }}{{ $dir := index $p 0 }}{{ $file := index $p 1 }}dir is {{$dir}}, file is {{$file}}'
dir is /tmp/, file is foo
```
