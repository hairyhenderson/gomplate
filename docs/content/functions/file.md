---
title: file functions
menu:
  main:
    parent: functions
---

## `file.Exists`

Reports whether a file or directory exists at the given path.

### Usage
```go
file.Exists path
```

### Example

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

### Usage
```go
file.IsDir path
```

### Example

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

Reads a given file _as text_. Note that this will succeed if the given file
is binary, but 

### Usage
```go
file.Read path
```

### Examples

```console
$ echo "hello world" > /tmp/hi
$ gomplate -i '{{file.Read "/tmp/hi"}}'
hello world
```

## `file.ReadDir`

Reads a directory and lists the files and directories contained within.

### Usage
```go
file.ReadDir path
```

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

Returns a [`os.FileInfo`](https://golang.org/pkg/os/#FileInfo) describing
the named path. 
Essentially a wrapper for Go's [`os.Stat`](https://golang.org/pkg/os/#Stat) function.

### Usage
```go
file.Stat path
```

### Examples

```console
$ echo "hello world" > /tmp/foo
$ gomplate -i '{{ $s := file.Stat "/tmp/foo" }}{{ $s.Mode }} {{ $s.Size }} {{ $s.Name }}'
-rw-r--r-- 12 foo
```
