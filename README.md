
# Gomplate

Flanksource Gomplate is a fork of [hairyhenderson/gomplate](https://github.com/hairyhenderson/gomplate) that provides a unified templating library supporting multiple expression languages.

## Features

- **Go Text/Template** – Full [Go `text/template`](https://pkg.go.dev/text/template) support with an extended function library from gomplate (base64, collections, crypto, data formats, filepath, math, random, regexp, strings, time, and more)
- **CEL (Common Expression Language)** – [CEL](https://cel.dev/) support with:
  - Standard CEL operators and built-ins
  - [cel-go extensions](https://github.com/cel-expr/cel-go/tree/v0.31.0/ext) (strings, encoders, lists, math, sets)
  - Kubernetes-specific helpers (`k8s.*`)
  - AWS helpers (`aws.*`)
  - Many gomplate functions remapped into CEL (`base64`, `math`, `random`, `regexp`, `filepath`, `crypto`, `sets`, etc.)
  - Optional-type support (`obj.?field.orValue("default")`)
- **JavaScript (otto)** – Lightweight JS evaluation via [otto](https://github.com/robertkrimen/otto)
- **Struct templating** – Walk Go structs/maps and apply templates to string fields in-place

## Quick Reference

| Language | Example | Documentation |
|---|---|---|
| Go Template | `{{ .name \| strings.ToUpper }}` | [GO_TEMPLATE.md](GO_TEMPLATE.md) |
| CEL | `name.upperAscii()` | [CEL.md](CEL.md) |

### CEL Quick Reference

```
# Arithmetic
2 + 3           // 5
"hello" + " world"   // "hello world"
[1,2] + [3,4]   // [1,2,3,4]

# Comparison / Logic
x > 5 && y != null
true ? "yes" : "no"

# String methods
"hello world".upperAscii()        // "HELLO WORLD"
"hello world".split(" ")          // ["hello", "world"]
"hello".startsWith("he")          // true
"hello world".contains("world")   // true

# Collections
[1,2,3].filter(e, e > 1)          // [2, 3]
[1,2,3].map(e, e * 2)             // [2, 4, 6]
[1,2,3].all(e, e > 0)             // true
{"a":1,"b":2}.keys()              // ["a","b"]

# Optional / null-safety
obj.?field.orValue("default")
obj.?a.?b.orValue("fallback")

# Kubernetes
k8s.isHealthy(pod)
k8s.cpuAsMillicores("500m")       // 500
k8s.memoryAsBytes("1Gi")

# Math
math.Add([1,2,3])                 // 6
math.greatest([1,2,3])            // 3

# Time
time.Now()
time.Since(timestamp("2023-01-01T00:00:00Z"))
```

See [CEL.md](CEL.md) for full reference.

### Go Template Quick Reference

```
# Variables and output
{{ .name }}
{{ $x := "hello" }}{{ $x | strings.ToUpper }}

# Control flow
{{ if eq .status "ok" }}OK{{ else }}FAIL{{ end }}
{{ range .items }}{{ . }}{{ end }}

# String functions
{{ "hello world" | strings.ToUpper }}        // HELLO WORLD
{{ "hello world" | strings.Split " " }}      // [hello world]
{{ "hello" | strings.Repeat 3 }}             // hellohellohello

# Data
{{ '{"key":"val"}' | json }}                 // map access
{{ .obj | toJSON }}
{{ .obj | toYAML }}

# Collections
{{ coll.Dict "a" 1 "b" 2 | toJSON }}         // {"a":1,"b":2}
{{ slice 1 2 3 | coll.Reverse | toJSON }}     // [3,2,1]
{{ coll.Sort (slice "b" "a" "c") | toJSON }}  // ["a","b","c"]

# Math
{{ math.Add 1 2 3 }}                          // 6
{{ math.Max 1 5 3 }}                          // 5

# base64
{{ base64.Encode "hello" }}                   // aGVsbG8=
{{ "aGVsbG8=" | base64.Decode }}              // hello

# Crypto
{{ crypto.SHA256 "hello" }}

# Delimiters (change if conflict with Helm)
# gotemplate: left-delim=$[[ right-delim=]]
$[[ .name ]]
```

See [GO_TEMPLATE.md](GO_TEMPLATE.md) for full reference.

## Installation

```go
import "github.com/flanksource/gomplate/v3"
```

## Usage

### Go Template

```go
result, err := gomplate.RunTemplate(map[string]any{
    "name": "world",
}, gomplate.Template{
    Template: `Hello, {{ .name }}!`,
})
```

### CEL

```go
result, err := gomplate.RunTemplate(map[string]any{
    "name": "world",
}, gomplate.Template{
    Expression: `"Hello, " + name + "!"`,
})
```

### Struct Templating

```go
type Config struct {
    Message string
    Count   int
}

cfg := Config{Message: "Hello, {{ .name }}!"}

err := gomplate.Walk(map[string]any{"name": "world"}, &cfg)
// cfg.Message == "Hello, world!"

## CEL eval helper

A small helper command is available at `cmd/ceval` to evaluate CEL expressions against an input YAML or JSON file.

Run it with:

```bash
go run ./cmd/ceval -f env.yaml -e "labels['a/b/c']"
```

Example input file:

```yaml
labels:
  a/b/c: d
```

Output:

```text
d
```
