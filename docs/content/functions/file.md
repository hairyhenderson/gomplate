---
title: file functions
menu:
  main:
    parent: functions
---

Functions for working with files.

## `file.Exists`

Reports whether a file or directory exists at the given path.

_Added in gomplate [v2.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.4.0)_
### Usage

```
file.Exists path
```
```
path | file.Exists
```

### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The path |

### Examples

_`input.tmpl`:_
```
{{ if (file.Exists "/tmp/foo") }}yes{{else}}no{{end}}
```

```console
$ gomplate -f input.tmpl
no
$ touch /tmp/foo
$ gomplate -f input.tmpl
yes
```

## `file.IsDir`

Reports whether a given path is a directory.

_Added in gomplate [v2.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.4.0)_
### Usage

```
file.IsDir path
```
```
path | file.IsDir
```

### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The path |

### Examples

_`input.tmpl`:_
```
{{ if (file.IsDir "/tmp/foo") }}yes{{else}}no{{end}}
```

```console
$ gomplate -f input.tmpl
no
$ touch /tmp/foo
$ gomplate -f input.tmpl
no
$ rm /tmp/foo && mkdir /tmp/foo
$ gomplate -f input.tmpl
yes
```

## `file.Read`

Reads a given file _as text_. Note that this will succeed if the given file is binary, but the output may be gibberish.

_Added in gomplate [v2.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.4.0)_
### Usage

```
file.Read path
```
```
path | file.Read
```

### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The path |

### Examples

```console
$ echo "hello world" > /tmp/hi
$ gomplate -i '{{file.Read "/tmp/hi"}}'
hello world
```

## `file.ReadDir`

Reads a directory and lists the files and directories contained within.

_Added in gomplate [v2.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.4.0)_
### Usage

```
file.ReadDir path
```
```
path | file.ReadDir
```

### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The path |

### Examples

```console
$ mkdir /tmp/foo
$ touch /tmp/foo/a; touch /tmp/foo/b; touch /tmp/foo/c
$ mkdir /tmp/foo/d
$ gomplate -i '{{ range (file.ReadDir "/tmp/foo") }}{{.}}{{"\n"}}{{end}}'
a
b
c
d
```

## `file.Stat`

Returns a [`os.FileInfo`](https://pkg.go.dev/os/#FileInfo) describing the named path.

Essentially a wrapper for Go's [`os.Stat`](https://pkg.go.dev/os/#Stat) function.

_Added in gomplate [v2.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.4.0)_
### Usage

```
file.Stat path
```
```
path | file.Stat
```

### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The path |

### Examples

```console
$ echo "hello world" > /tmp/foo
$ gomplate -i '{{ $s := file.Stat "/tmp/foo" }}{{ $s.Mode }} {{ $s.Size }} {{ $s.Name }}'
-rw-r--r-- 12 foo
```

## `file.Walk`

Like a recursive [`file.ReadDir`](#filereaddir), recursively walks the file tree rooted at `path`, and returns an array of all files and directories contained within.

The files are walked in lexical order, which makes the output deterministic but means that for very large directories can be inefficient.

Walk does not follow symbolic links.

Similar to Go's [`filepath.Walk`](https://pkg.go.dev/path/filepath/#Walk) function.

_Added in gomplate [v2.6.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.6.0)_
### Usage

```
file.Walk path
```
```
path | file.Walk
```

### Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The path |

### Examples

```console
$ tree /tmp/foo
/tmp/foo
├── one
├── sub
│   ├── one
│   └── two
├── three
└── two

1 directory, 5 files
$ gomplate -i '{{ range file.Walk "/tmp/foo" }}{{ if not (file.IsDir .) }}{{.}} is a file{{"\n"}}{{end}}{{end}}'
/tmp/foo/one is a file
/tmp/foo/sub/one is a file
/tmp/foo/sub/two is a file
/tmp/foo/three is a file
/tmp/foo/two is a file
```

## `file.Write`

Write the given data to the given file. If the file exists, it will be overwritten.

For increased security, `file.Write` will only write to files which are contained within the current working directory. Attempts to write elsewhere will fail with an error.

Non-existing directories in the output path will be created.

If the data is a byte array (`[]byte`), it will be written as-is. Otherwise, it will be converted to a string before being written.

_Added in gomplate [v2.4.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.4.0)_
### Usage

```
file.Write filename data
```
```
data | file.Write filename
```

### Arguments

| name | description |
|------|-------------|
| `filename` | _(required)_ The name of the file to write to |
| `data` | _(required)_ The data to write |

### Examples

```console
$ gomplate -i '{{ file.Write "/tmp/foo" "hello world" }}'
$ cat /tmp/foo
hello world
```
