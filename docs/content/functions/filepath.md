---
title: filepath functions
menu:
  main:
    parent: functions
---

gomplate's path functions are split into 2 namespaces:
- `path`, which is useful for manipulating slash-based (`/`) paths, such as in URLs
- `filepath`, which should be used for local filesystem paths, especially when Windows paths may be involved.

This page documents the `filepath` namespace - see also the [`path`](../path) documentation.

These functions are wrappers for Go's [`path/filepath`](https://pkg.go.dev/path/filepath/) package.

## `filepath.Base`

Returns the last element of path. Trailing path separators are removed before extracting the last element. If the path is empty, Base returns `.`. If the path consists entirely of separators, Base returns a single separator.

A wrapper for Go's [`filepath.Base`](https://pkg.go.dev/path/filepath/#Base) function.

_Added in gomplate [v2.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.7.0)_
### Usage

```
filepath.Base path
```
```
path | filepath.Base
```

### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

### Examples

```console
$ gomplate -i '{{ filepath.Base "/tmp/foo" }}'
foo
```

## `filepath.Clean`

Clean returns the shortest path name equivalent to path by purely lexical processing.

A wrapper for Go's [`filepath.Clean`](https://pkg.go.dev/path/filepath/#Clean) function.

_Added in gomplate [v2.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.7.0)_
### Usage

```
filepath.Clean path
```
```
path | filepath.Clean
```

### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

### Examples

```console
$ gomplate -i '{{ filepath.Clean "/tmp//foo/../" }}'
/tmp
```

## `filepath.Dir`

Returns all but the last element of path, typically the path's directory.

A wrapper for Go's [`filepath.Dir`](https://pkg.go.dev/path/filepath/#Dir) function.

_Added in gomplate [v2.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.7.0)_
### Usage

```
filepath.Dir path
```
```
path | filepath.Dir
```

### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

### Examples

```console
$ gomplate -i '{{ filepath.Dir "/tmp/foo" }}'
/tmp
```

## `filepath.Ext`

Returns the file name extension used by path.

A wrapper for Go's [`filepath.Ext`](https://pkg.go.dev/path/filepath/#Ext) function.

_Added in gomplate [v2.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.7.0)_
### Usage

```
filepath.Ext path
```
```
path | filepath.Ext
```

### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

### Examples

```console
$ gomplate -i '{{ filepath.Ext "/tmp/foo.csv" }}'
.csv
```

## `filepath.FromSlash`

Returns the result of replacing each slash (`/`) character in the path with the platform's separator character.

A wrapper for Go's [`filepath.FromSlash`](https://pkg.go.dev/path/filepath/#FromSlash) function.

_Added in gomplate [v2.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.7.0)_
### Usage

```
filepath.FromSlash path
```
```
path | filepath.FromSlash
```

### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

### Examples

```console
$ gomplate -i '{{ filepath.FromSlash "/foo/bar" }}'
/foo/bar
C:\> gomplate.exe -i '{{ filepath.FromSlash "/foo/bar" }}'
C:\foo\bar
```

## `filepath.IsAbs`

Reports whether the path is absolute.

A wrapper for Go's [`filepath.IsAbs`](https://pkg.go.dev/path/filepath/#IsAbs) function.

_Added in gomplate [v2.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.7.0)_
### Usage

```
filepath.IsAbs path
```
```
path | filepath.IsAbs
```

### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

### Examples

```console
$ gomplate -i 'the path is {{ if (filepath.IsAbs "/tmp/foo.csv") }}absolute{{else}}relative{{end}}'
the path is absolute
$ gomplate -i 'the path is {{ if (filepath.IsAbs "../foo.csv") }}absolute{{else}}relative{{end}}'
the path is relative
```

## `filepath.Join`

Joins any number of path elements into a single path, adding a separator if necessary.

A wrapper for Go's [`filepath.Join`](https://pkg.go.dev/path/filepath/#Join) function.

_Added in gomplate [v2.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.7.0)_
### Usage

```
filepath.Join elem...
```

### Arguments

| name | description |
|------|-------------|
| `elem...` | _(required)_ The path elements to join (0 or more) |

### Examples

```console
$ gomplate -i '{{ filepath.Join "/tmp" "foo" "bar" }}'
/tmp/foo/bar
C:\> gomplate.exe -i '{{ filepath.Join "C:\tmp" "foo" "bar" }}'
C:\tmp\foo\bar
```

## `filepath.Match`

Reports whether name matches the shell file name pattern.

A wrapper for Go's [`filepath.Match`](https://pkg.go.dev/path/filepath/#Match) function.

_Added in gomplate [v2.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.7.0)_
### Usage

```
filepath.Match pattern path
```

### Arguments

| name | description |
|------|-------------|
| `pattern` | _(required)_ The pattern to match on |
| `path` | _(required)_ The path to match |

### Examples

```console
$ gomplate -i '{{ filepath.Match "*.csv" "foo.csv" }}'
true
```

## `filepath.Rel`

Returns a relative path that is lexically equivalent to targetpath when joined to basepath with an intervening separator.

A wrapper for Go's [`filepath.Rel`](https://pkg.go.dev/path/filepath/#Rel) function.

_Added in gomplate [v2.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.7.0)_
### Usage

```
filepath.Rel basepath targetpath
```

### Arguments

| name | description |
|------|-------------|
| `basepath` | _(required)_ The base path |
| `targetpath` | _(required)_ The target path |

### Examples

```console
$ gomplate -i '{{ filepath.Rel "/a" "/a/b/c" }}'
b/c
```

## `filepath.Split`

Splits path immediately following the final path separator, separating it into a directory and file name component.

The function returns an array with two values, the first being the diretory, and the second the file.

A wrapper for Go's [`filepath.Split`](https://pkg.go.dev/path/filepath/#Split) function.

_Added in gomplate [v2.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.7.0)_
### Usage

```
filepath.Split path
```
```
path | filepath.Split
```

### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

### Examples

```console
$ gomplate -i '{{ $p := filepath.Split "/tmp/foo" }}{{ $dir := index $p 0 }}{{ $file := index $p 1 }}dir is {{$dir}}, file is {{$file}}'
dir is /tmp/, file is foo
C:\> gomplate.exe -i '{{ $p := filepath.Split `C:\tmp\foo` }}{{ $dir := index $p 0 }}{{ $file := index $p 1 }}dir is {{$dir}}, file is {{$file}}'
dir is C:\tmp\, file is foo
```

## `filepath.ToSlash`

Returns the result of replacing each separator character in path with a slash (`/`) character.

A wrapper for Go's [`filepath.ToSlash`](https://pkg.go.dev/path/filepath/#ToSlash) function.

_Added in gomplate [v2.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.7.0)_
### Usage

```
filepath.ToSlash path
```
```
path | filepath.ToSlash
```

### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

### Examples

```console
$ gomplate -i '{{ filepath.ToSlash "/foo/bar" }}'
/foo/bar
C:\> gomplate.exe -i '{{ filepath.ToSlash `foo\bar\baz` }}'
foo/bar/baz
```

## `filepath.VolumeName`

Returns the leading volume name. Given `C:\foo\bar` it returns `C:` on Windows. Given a UNC like `\\host\share\foo` it returns `\\host\share`. On other platforms it returns an empty string.

A wrapper for Go's [`filepath.VolumeName`](https://pkg.go.dev/path/filepath/#VolumeName) function.

_Added in gomplate [v2.7.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.7.0)_
### Usage

```
filepath.VolumeName path
```
```
path | filepath.VolumeName
```

### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

### Examples

```console
C:\> gomplate.exe -i 'volume is {{ filepath.VolumeName "C:/foo/bar" }}'
volume is C:
$ gomplate -i 'volume is {{ filepath.VolumeName "/foo/bar" }}'
volume is
```
