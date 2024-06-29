---
title: Functions
---

Almost all of gomplate's utility is provided as _functions._ These are key
words that perform some action.

For example, the [`base64.Encode`][] function will encode some input string as
a base-64 string:

```
The word is {{ base64.Encode "swordfish" }}
```
renders as:
```
The word is c3dvcmRmaXNo
```

## Built-ins

Go's [`text/template`][] language provides a number of built-in functions, 
operators, and actions that can be used in templates.

### Built-in functions

Here is a list of the built-in functions, but see [the documentation](https://golang.org/pkg/text/template/#hdr-Functions)
for full details:

- `and`, `or`, `not`: Returns boolean AND/OR/NOT of the argument(s).
- `call`: Returns the result of calling a function argument.
- `html`, `js`, `urlquery`: Safely escapes input for inclusion in HTML, JavaScript, and URL query strings.
- `index`: Returns the referenced element of an array/slice, string, or map. See also [Arrays](../syntax/#arrays) and [Maps](../syntax/#maps).
- `len`: Returns the length of the argument.
- `print`, `printf`, `println`: Aliases for Go's [`fmt.Print`](https://golang.org/pkg/fmt/#Print),
[`fmt.Printf`](https://golang.org/pkg/fmt/#Printf), and [`fmt.Println`](https://golang.org/pkg/fmt/#Println)
functions. See the [format documentation](https://golang.org/pkg/fmt/#hdr-Printing)
for details on `printf`'s format syntax.

### Operators

And the following comparison operators are also supported:

- `eq`: Equal (`==`)
- `ne`: Not-equal (`!=`)
- `lt`: Less than (`<`)
- `le`: Less than or equal to (`<=`)
- `gt`: Greater than (`>`)
- `ge`: Greater than or equal to (`>=`)

### Actions

There are also a few _actions_, which are used for control flow and other purposes. See [the documentation](https://golang.org/pkg/text/template/#hdr-Actions) for details on these:

- `if`/`else`/`else if`: Conditional control flow.
- `with`/`else`: Conditional execution with assignment.
- `range`: Looping control flow. See discussion in the [Arrays](../syntax/#arrays) and [Maps](../syntax/#maps) sections.
  - `break`: The innermost `range` loop is ended early, stopping the current iteration and bypassing all remaining iterations.
  - `continue`: The current iteration of the innermost `range` loop is stopped, and the loop starts the next iteration.
- `template`: Include the output of a named template. See the [Nested templates](/syntax/#nested-templates) section for more details, and the [`tmpl`](../functions/tmpl) namespace for more flexible versions of `template`.
- `define`: Define a named nested template. See the [Nested templates](/syntax/#nested-templates) section for more details.
- `block`: Shorthand for `define` followed immediately by `template`.

## gomplate functions

gomplate provides over 200 functions not found in the standard library. These
are grouped into namespaces, and documented on the following pages:

{{% children depth="3" description="false" %}}

[`text/template`]: https://golang.org/pkg/text/template/
[`base64.Encode`]: ../functions/base64#base64encode
[data sources]: ../datasources/
