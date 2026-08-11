# Feature: nil-safe `first` and `last` — CEL and gomplate functions

## Overview

Add `first` and `last` functions available as CEL global functions, CEL member
methods, and gomplate template functions. They return the first/last element of
a list, character of a string, or value (by sorted-key order) of a map, with
nil-safe behaviour: null/empty/out-of-range inputs return the zero value of the
element type rather than panicking or producing an error.

**Problem**: CEL expressions frequently need to pick the head or tail of a
collection produced by a nilsafe field access (which may be null or empty). No
such primitive exists in the codebase today, so users resort to verbose
conditionals.

**Target users**: Engineers writing CEL expressions and Go-template strings in
Mission Control config/playbooks.

---

## Functional Requirements

### FR-1: `first` — return first element of a list

**Description**: Returns the element at index 0 of a list/slice. Returns the
zero value of the element type when the list is null or empty.

**User Story**: As a CEL user, I want `first(items)` or `items.first()` so that
I can safely extract the leading element without guarding against empty lists.

**Acceptance Criteria**:
- [ ] `first([1, 2, 3])` → `1`
- [ ] `first([])` → `0` (int zero) / `""` (string zero) / `null` when type unknown
- [ ] `first(null)` → `null`
- [ ] Available as `first(list)` (global) and `list.first()` (member) in CEL
- [ ] Available as `{{ first .list }}` in gomplate templates

---

### FR-2: `last` — return last element of a list

**Description**: Returns the element at index `len-1` of a list/slice. Returns
the zero value of the element type when the list is null or empty.

**User Story**: As a CEL user, I want `last(items)` or `items.last()` so that I
can safely extract the trailing element.

**Acceptance Criteria**:
- [ ] `last([1, 2, 3])` → `3`
- [ ] `last([])` → zero value / `null`
- [ ] `last(null)` → `null`
- [ ] Available as `last(list)` (global) and `list.last()` (member) in CEL
- [ ] Available as `{{ last .list }}` in gomplate

---

### FR-3: String support — first/last character

**Description**: When the input is a string, `first` returns the first character
(as a single-character string) and `last` returns the last character.

**Acceptance Criteria**:
- [ ] `first("hello")` → `"h"`
- [ ] `last("hello")` → `"o"`
- [ ] `first("")` → `""` (empty string zero value)
- [ ] `last("")` → `""`
- [ ] `first(null)` (string-typed null) → `""`

---

### FR-4: Map support — first/last value by sorted key

**Description**: When the input is a map, `first` returns the value at the
lexicographically smallest key, and `last` returns the value at the largest key.

**Acceptance Criteria**:
- [ ] `first({"b": 2, "a": 1})` → `1` (key "a" sorts first)
- [ ] `last({"b": 2, "a": 1})` → `2` (key "b" sorts last)
- [ ] `first({})` → `null`
- [ ] `last({})` → `null`
- [ ] `first(null)` (map-typed null) → `null`

---

### FR-5: Nil-safe — no errors on null/empty input

**Description**: Both functions must never return a CEL error or a Go panic for
null, empty, or out-of-range inputs. They should be exempt from the `nilsafe`
library's short-circuit decorator (same pattern as `coalesce`).

**Acceptance Criteria**:
- [ ] Exempt from `nilSafeCall` decorator in `nilsafe/nilsafe.go`
- [ ] Null list argument → returns `null` (not an error)
- [ ] Empty list argument → returns type-matched zero value or `null`
- [ ] Integrates cleanly with nilsafe variable resolution (`x.missing.first()` → `""` / `null`)

---

## Technical Considerations

### Implementation location

| Artifact | Location |
|----------|----------|
| Pure Go logic | `coll/firstlast.go` (new file) |
| gomplate wrapper | `funcs/coll.go` — add `First` / `Last` methods to `CollFuncs`, register in `CreateCollFuncs` |
| CEL global bindings | `funcs/coll.go` — hand-written `cel.Function` vars `celFirst`, `celLast` |
| CEL member bindings | `funcs/coll.go` — `cel.MemberOverload` variants for list, string, map |
| CEL registration | `funcs/cel_exports.go` — append `celFirst`, `celLast` |
| nilsafe exemption | `nilsafe/nilsafe.go` — add `"first"` and `"last"` to the bypass check (same as `"coalesce"`) |
| Tests (unit) | `coll/firstlast_test.go` |
| Tests (CEL integration) | `tests/cel_test.go` — `TestCelFirstLast` |

### CEL overload strategy

CEL-go does not support true variadic functions but does support member
overloads. Register both global and member forms:

```go
// Global: first(list)
cel.Overload("first_list", []*cel.Type{cel.ListType(cel.DynType)}, cel.DynType, ...)
cel.Overload("first_string", []*cel.Type{cel.StringType}, cel.StringType, ...)
cel.Overload("first_map", []*cel.Type{cel.MapType(cel.StringType, cel.DynType)}, cel.DynType, ...)

// Member: list.first()
cel.MemberOverload("list_first", []*cel.Type{cel.ListType(cel.DynType)}, cel.DynType, ...)
cel.MemberOverload("string_first", []*cel.Type{cel.StringType}, cel.StringType, ...)
cel.MemberOverload("map_first", []*cel.Type{cel.MapType(cel.StringType, cel.DynType)}, cel.DynType, ...)
```

### Zero value strategy

Return type-matched zero values following the same pattern as `nilsafe/zeroval.go`:

| Input type | Empty/null result |
|------------|------------------|
| `list<string>` | `""` |
| `list<int>` | `0` |
| `list<bool>` | `false` |
| `list<dyn>` / unknown | `null` |
| `string` | `""` |
| `map` (empty/null) | `null` |

In practice, since CEL lists are typed `dyn` at runtime, use `types.NullValue`
for empty lists unless the first element's type can be inferred.

### Go / gomplate side

```go
func First(in any) any {
    // reflect-based: handles []T, string, map[string]any
}
func Last(in any) any { ... }
```

Nil/empty input returns `nil` (template renders as `""`).

---

## Success Criteria

- [ ] `first` and `last` are callable in CEL as global functions
- [ ] `first` and `last` are callable in CEL as member methods on list, string, map
- [ ] `first` and `last` are callable in gomplate templates
- [ ] Null/empty inputs never produce CEL errors or Go panics
- [ ] Nil-safe integration: `x.missing.first()` works when `x.missing` is null
- [ ] All unit and integration tests pass (`go test ./...`)
- [ ] `make lint` introduces no new issues

---

## Testing Requirements

### Unit tests — `coll/firstlast_test.go`

| Case | Input | Fn | Expected |
|------|-------|----|----------|
| list first element | `[]any{1, 2, 3}` | First | `1` |
| list last element | `[]any{1, 2, 3}` | Last | `3` |
| single-element list | `[]any{"x"}` | First/Last | `"x"` |
| empty list | `[]any{}` | First/Last | `nil` |
| nil input | `nil` | First/Last | `nil` |
| string first char | `"hello"` | First | `"h"` |
| string last char | `"hello"` | Last | `"o"` |
| empty string | `""` | First/Last | `""` |
| map first by key | `map[string]any{"b":2,"a":1}` | First | `1` |
| map last by key | `map[string]any{"b":2,"a":1}` | Last | `2` |
| empty map | `map[string]any{}` | First/Last | `nil` |

### CEL integration tests — `tests/cel_test.go` (`TestCelFirstLast`)

| Expression | Env | Expected |
|------------|-----|----------|
| `first([1,2,3])` | — | `"1"` |
| `last([1,2,3])` | — | `"3"` |
| `[1,2,3].first()` | — | `"1"` |
| `[1,2,3].last()` | — | `"3"` |
| `first([])` | — | `""` |
| `last([])` | — | `""` |
| `first(null)` | — | `""` |
| `first("hello")` | — | `"h"` |
| `last("hello")` | — | `"o"` |
| `"hello".first()` | — | `"h"` |
| `first({"b":2,"a":1})` | — | `"1"` |
| `last({"b":2,"a":1})` | — | `"2"` |
| `first(a)` | `a=null` | `""` |
| nil-safe: `a.first()` | `a=null (list)` | `""` |

---

## Implementation Checklist

### Phase 1: Pure Go logic
- [ ] Create `coll/firstlast.go` with `First(in any) any` and `Last(in any) any`
- [ ] Write `coll/firstlast_test.go` covering all input types and edge cases
- [ ] Verify: `go test ./coll/...`

### Phase 2: gomplate template functions
- [ ] Add `First` / `Last` methods to `CollFuncs` in `funcs/coll.go`
- [ ] Register `"first"` and `"last"` keys in `CreateCollFuncs`

### Phase 3: CEL bindings
- [ ] Write `celFirst` / `celLast` `cel.Function` vars with global + member overloads for list, string, map
- [ ] Add `"first"` and `"last"` to the nilsafe bypass in `nilsafe/nilsafe.go`
- [ ] Append `celFirst`, `celLast` to `CelEnvOption` in `funcs/cel_exports.go`

### Phase 4: Integration tests
- [ ] Add `TestCelFirstLast` to `tests/cel_test.go`
- [ ] Run full suite: `go test ./...`
- [ ] Confirm no new lint issues: `make lint`
