---
title: strings functions
menu:
  main:
    parent: functions
---

## `strings.Contains`

Reports whether a substring is contained within a string.

### Usage
```go
strings.Contains substr input
```
```go
input | strings.Contains substr
```

### Example

_`input.tmpl`:_
```
{{ if (.Env.FOO | strings.Contains "f") }}yes{{else}}no{{end}}
```

```console
$ FOO=foo gomplate < input.tmpl
yes
$ FOO=bar gomplate < input.tmpl
no
```

## `strings.HasPrefix`

Tests whether a string begins with a certain prefix.

### Usage
```go
strings.HasPrefix prefix input
```
```go
input | strings.HasPrefix prefix
```

#### Example

```console
$ URL=http://example.com gomplate -i '{{if .Env.URL | strings.HasPrefix "https"}}foo{{else}}bar{{end}}'
bar
$ URL=https://example.com gomplate -i '{{if .Env.URL | strings.HasPrefix "https"}}foo{{else}}bar{{end}}'
foo
```

## `strings.HasSuffix`

Tests whether a string ends with a certain suffix.

### Usage
```go
strings.HasSuffix suffix input
```
```go
input | strings.HasSuffix suffix
```

### Examples

_`input.tmpl`:_
```
{{.Env.URL}}{{if not (.Env.URL | strings.HasSuffix ":80")}}:80{{end}}
```

```console
$ URL=http://example.com gomplate < input.tmpl
http://example.com:80
```

## `strings.Split`

Creates a slice by splitting a string on a given delimiter.

### Usage
```go
strings.Split separator input
```
```go
input | strings.Split separator
```

### Examples

```console
$ gomplate -i '{{range ("Bart,Lisa,Maggie" | strings.Split ",") }}Hello, {{.}}{{end}}'
Hello, Bart
Hello, Lisa
Hello, Maggie
```

## `strings.SplitN`

Creates a slice by splitting a string on a given delimiter. The count determines
the number of substrings to return.

### Usage
```go
strings.SplitN separator count input
```
```go
input | strings.SplitN separator count
```

#### Example

```console
$ gomplate -i '{{ range ("foo:bar:baz" | strings.SplitN ":" 2) }}{{.}}{{end}}'
foo
bar:baz
```

## `strings.ReplaceAll`

**Alias:** `replaceAll`

Replaces all occurrences of a given string with another.

### Usage
```go
strings.ReplaceAll old new input
```
```go
input | strings.ReplaceAll old new
```

### Examples

```console
$ gomplate -i '{{ strings.ReplaceAll "." "-" "172.21.1.42" }}'
172-21-1-42
$ gomplate -i '{{ "172.21.1.42" | strings.ReplaceAll "." "-" }}'
172-21-1-42
```

## `strings.Title`

**Alias:** `title`

Convert to title-case.

### Usage
```go
strings.Title input
```
```go
input | strings.Title
```

### Example

```console
$ gomplate -i '{{strings.Title "hello, world!"}}'
Hello, World!
```

## `strings.ToLower`

**Alias:** `toLower`

Convert to lower-case.

### Usage
```go
strings.ToLower input
```
```go
input | strings.ToLower
```

#### Example

```console
$ echo '{{strings.ToLower "HELLO, WORLD!"}}' | gomplate
hello, world!
```

## `strings.ToUpper`

**Alias:** `toUpper`

Convert to upper-case.

### Usage
```go
strings.ToUpper input
```
```go
input | strings.ToUpper
```

#### Example

```console
$ gomplate -i '{{strings.ToUpper "hello, world!"}}'
HELLO, WORLD!
```

## `strings.Trim`

Trims a string by removing the given characters from the beginning and end of
the string.

### Usage
```go
strings.Trim cutset input
```
```go
input | strings.Trim cutset
```

#### Example

```console
$ gomplate -i '{{ "_-foo-_" | strings.Trim "_-" }}
foo
```

## `strings.TrimSpace`

**Alias:** `trimSpace`

Trims a string by removing whitespace from the beginning and end of
the string.

### Usage
```go
strings.TrimSpace input
```
```go
input | strings.TrimSpace
```

#### Example

```console
$ gomplate -i '{{ "  \n\t foo" | strings.TrimSpace }}'
foo
```

## `contains`

**See [`strings.Contains](#strings-contains) for a pipeline-compatible version**

Contains reports whether the second string is contained within the first. Equivalent to
[strings.Contains](https://golang.org/pkg/strings#Contains)

### Example

_`input.tmpl`:_
```
{{if contains .Env.FOO "f"}}yes{{else}}no{{end}}
```

```console
$ FOO=foo gomplate < input.tmpl
yes
$ FOO=bar gomplate < input.tmpl
no
```

## `hasPrefix`

**See [`strings.HasPrefix](#strings-hasprefix) for a pipeline-compatible version**

Tests whether the string begins with a certain substring. Equivalent to
[strings.HasPrefix](https://golang.org/pkg/strings#HasPrefix)

#### Example

_`input.tmpl`:_
```
{{if hasPrefix .Env.URL "https"}}foo{{else}}bar{{end}}
```

```console
$ URL=http://example.com gomplate < input.tmpl
bar
$ URL=https://example.com gomplate < input.tmpl
foo
```

## `hasSuffix`

**See [`strings.HasSuffix](#strings-hassuffix) for a pipeline-compatible version**

Tests whether the string ends with a certain substring. Equivalent to
[strings.HasSuffix](https://golang.org/pkg/strings#HasSuffix)

#### Example

_`input.tmpl`:_
```
{{.Env.URL}}{{if not (hasSuffix .Env.URL ":80")}}:80{{end}}
```

```console
$ URL=http://example.com gomplate < input.tmpl
http://example.com:80
```

## `split`

**See [`strings.Split](#strings-split) for a pipeline-compatible version**

Creates a slice by splitting a string on a given delimiter. Equivalent to
[strings.Split](https://golang.org/pkg/strings#Split)

#### Example

```console
$ gomplate -i '{{range split "Bart,Lisa,Maggie" ","}}Hello, {{.}}{{end}}'
Hello, Bart
Hello, Lisa
Hello, Maggie
```

## `splitN`

**See [`strings.SplitN](#strings-splitn) for a pipeline-compatible version**

Creates a slice by splitting a string on a given delimiter. The count determines
the number of substrings to return. Equivalent to [strings.SplitN](https://golang.org/pkg/strings#SplitN)

#### Example

```console
$ gomplate -i '{{ range splitN "foo:bar:baz" ":" 2 }}{{.}}{{end}}'
foo
bar:baz
```

## `trim`

**See [`strings.Trim](#strings-trim) for a pipeline-compatible version**

Trims a string by removing the given characters from the beginning and end of
the string. Equivalent to [strings.Trim](https://golang.org/pkg/strings/#Trim)

#### Example

_`input.tmpl`:_
```
Hello, {{trim .Env.FOO " "}}!
```

```console
$ FOO="  world " | gomplate < input.tmpl
Hello, world!
```
