# CEL Expressions

CEL expressions use the [Common Expression Language (CEL)](https://cel.dev/).

> **Tip:** The [CEL playground](https://playcel.undistro.io/) lets you test CEL expressions.

## Types

| Type | Description |
|---|---|
| `int` | 64-bit signed integers |
| `uint` | 64-bit unsigned integers |
| `double` | 64-bit IEEE floating-point numbers |
| `bool` | Booleans (`true` or `false`) |
| `string` | Strings of Unicode code points |
| `bytes` | Byte sequences |
| `list` | Lists of values |
| `map` | Associative arrays with `int`, `uint`, `bool`, or `string` keys |
| `null_type` | The value `null` |
| `type` | Values representing the types above |

### Native Go struct types

Register a Go struct type to expose it directly to CEL instead of converting top-level values of that type to maps:

```go
type Person struct {
    DisplayName string `json:"display_name"`
}

if err := gomplate.RegisterType(Person{}); err != nil {
    return err
}

result, err := gomplate.RunExpression(
    map[string]any{"person": Person{DisplayName: "Ada"}},
    gomplate.Template{Expression: "person.display_name"},
)
```

Registration is process-wide, concurrency-safe, and applies to subsequent CEL compilations. JSON field names are honored; a JSON tag containing only options, such as `json:",omitempty"`, retains the Go field name. When an expression returns a registered top-level value directly, the result is the original Go value. Unregistered values and Go-template evaluation retain the existing serialization behavior.

---

## Standard Operators

### Arithmetic Operators

| Operator | Description | Example |
|---|---|---|
| `+` | Addition (also string/list concatenation) | `2 + 3` → `5`<br>`"hello" + " world"` → `"hello world"`<br>`[1, 2] + [3, 4]` → `[1, 2, 3, 4]` |
| `-` | Subtraction (also negation) | `5 - 3` → `2`<br>`-5` → `-5` |
| `*` | Multiplication | `3 * 4` → `12` |
| `/` | Division | `10 / 2` → `5` |
| `%` | Remainder (modulo) | `10 % 3` → `1` |

### Comparison Operators

| Operator | Description | Example |
|---|---|---|
| `==` | Equal | `5 == 5` → `true` |
| `!=` | Not equal | `5 != 3` → `true` |
| `<` | Less than | `3 < 5` → `true` |
| `<=` | Less than or equal | `5 <= 5` → `true` |
| `>` | Greater than | `5 > 3` → `true` |
| `>=` | Greater than or equal | `5 >= 5` → `true` |

### Logical Operators

| Operator | Description | Example |
|---|---|---|
| `&&` | Logical AND | `true && false` → `false` |
| `\|\|` | Logical OR | `true \|\| false` → `true` |
| `!` | Logical NOT | `!true` → `false` |
| `? :` | Ternary conditional | `true ? "yes" : "no"` → `"yes"` |

---

## Type Conversion Functions

| Function | Description | Example |
|---|---|---|
| `bool()` | Convert to boolean | `bool("true")` → `true` |
| `bytes()` | Convert to bytes | `bytes("hello")` → `b'hello'` |
| `double()` | Convert to double | `double(5)` → `5.0` |
| `duration()` | Convert to duration | `duration("1h")` → 1 hour duration |
| `int()` | Convert to integer | `int(5.7)` → `5` |
| `string()` | Convert to string | `string(123)` → `"123"` |
| `timestamp()` | Convert to timestamp | `timestamp("2023-01-01T00:00:00Z")` |
| `uint()` | Convert to unsigned integer | `uint(5)` → `5u` |
| `type()` | Get the type of a value | `type(5)` → `int` |
| `dyn()` | Create a dynamic value | `dyn({"key": "value"})` |

---

## Built-in Functions

### Type Checking

```
type(5)              // "int"
type("hello")        // "string"
type([1, 2, 3])      // "list"
type({"key": "value"}) // "map"
```

---

## Handling null types and missing keys

When dealing with CEL objects, a key might not exist or a middle key in a chain might be missing.

```
// Assume obj = {'a': {'b': 'c'}}
obj.a.b             // "c"
obj.a.d             // Error: attribute 'd' doesn't exist
obj.a.?d.orValue("fallback") // "fallback value"
```

See [or](#or) and [orValue](#orvalue) below for details.

---

## matchQuery

`matchQuery` matches a given resource against a search query.

```
matchQuery(r, s)
// r = resource
// s = search query

matchQuery(.config, "type=Kubernetes::Pod")
matchQuery(.config, "type=Kubernetes::Pod tags.cluster=homelab")
```

---

## matchLabel

`matchLabel` matches a map's key against one or more patterns. Useful for matching Kubernetes labels.

```
matchLabel(labels, key, patterns)
// labels  = map of labels
// key     = the label key to check
// patterns = comma-separated patterns to match against
```

**Pattern Syntax:**
- Use `*` for wildcards (e.g., `us-*` matches `us-east-1`, `us-west-2`)
- Use `!` for exclusion (e.g., `!production` matches any value except `production`)
- Use `!*` to match when the label doesn't exist
- Multiple patterns are evaluated as OR conditions

```
matchLabel(config.labels, "region", "us-*")        // true if region starts with "us-"
matchLabel(config.labels, "env", "prod,staging")   // true if env is "prod" OR "staging"
matchLabel(config.labels, "env", "!production")    // true if env is NOT "production"
matchLabel(config.labels, "optional", "!*")        // true if "optional" label doesn't exist
matchLabel(config.tags, "cluster", "*-prod,*-staging")
```

---

## aws

### aws.arnToMap

Takes in an AWS ARN, parses it, and returns a map.

```
aws.arnToMap("arn:aws:sns:eu-west-1:123:MMS-Topic")
// map[string]string{
//   "service": string,
//   "region": string,
//   "account": string,
//   "resource": string,
// }
```

### aws.fromAWSMap

`aws.fromAWSMap` takes a list of `map[string]string` and merges them into a single map. The input map has the field "Name".

```
aws.fromAWSMap(x).hello == "world"  // true
// Where x = [
//   { Name: 'hello', Value: 'world' },
//   { Name: 'John', Value: 'Doe' },
// ]
```

---

## base64

### base64.encode

Encodes the given byte slice to a Base64 encoded string.

```
base64.encode("hello") // "aGVsbG8="
```

### base64.decode

Decodes the given base64 encoded string back to its original form.

```
base64.decode("aGVsbG8=") // b'hello'
```

---

## collections

### .keys

Returns a list of keys from a map.

```
{"first": "John", "last": "Doe"}.keys()   // ["first", "last"]
```

### .merge

Merges a second map into the first.

```
{"first": "John"}.merge({"last": "Doe"})  // {"first": "John", "last": "Doe"}
```

### .omit

Removes a list of keys from a map.

```
{"first": "John", "last": "Doe"}.omit(["first"])  // {"last": "Doe"}
```

### .sort

Returns a sorted list.

```
[3, 2, 1].sort()           // [1, 2, 3]
['c', 'b', 'a'].sort()     // ['a', 'b', 'c']
```

### .distinct

Returns a new list with duplicate elements removed.

```
[1, 2, 2, 3, 3, 3].distinct()          // [1, 2, 3]
["a", "b", "a", "c"].distinct()        // ["a", "b", "c"]
```

### .flatten

Recursively flattens nested lists.

```
[[1, 2], [3, 4]].flatten()     // [1, 2, 3, 4]
[1, [2, [3, 4]]].flatten()     // [1, 2, 3, 4]
```

### .reverse

Returns a new list with elements in reverse order.

```
[1, 2, 3, 4].reverse()         // [4, 3, 2, 1]
["a", "b", "c"].reverse()      // ["c", "b", "a"]
```

### range

Generates a list of integers.

```
range(5)          // [0, 1, 2, 3, 4]
range(2, 5)       // [2, 3, 4]
range(0, 10, 2)   // [0, 2, 4, 6, 8]
```

### .uniq

Returns a list of unique items.

```
[1,2,3,3,3].uniq().sum()     // 10 (note: .sum() not available, illustrative)
["a", "b", "b"].uniq()       // ["a", "b"]
```

### .values

Returns a list of values from a map.

```
{'a': 1, 'b': 2}.values()    // [1, 2]
```

### keyValToMap

Converts a string in `key=value,key2=value2` format into a map.

```
keyValToMap("a=b,c=d")                      // {"a": "b", "c": "d"}
keyValToMap("env=prod,region=us-east-1")    // {"env": "prod", "region": "us-east-1"}
```

### mapToKeyVal

Converts a map into a string in `key=value,key2=value2` format.

```
{"a": "b", "c": "d"}.mapToKeyVal()         // "a=b,c=d"
```

### all

Tests whether a predicate holds for **all** elements of a list or keys of a map.

```
[1, 2, 3].all(e, e > 0)                                         // true
{"a": "apple", "b": "banana"}.all(k, k.startsWith("a"))         // false
```

### exists

Checks if there is at least one element in a list that satisfies a condition.

```
[1, 2, 3].exists(e, e == 2)     // true
```

### exists_one

Checks if there is exactly one element in a list that satisfies a condition.

```
[1, 2, 3].exists_one(e, e > 1)  // false
[1, 2, 3].exists_one(e, e == 2) // true
```

### filter

Creates a new list containing only elements that satisfy a condition.

```
[1, 2, 3, 4].filter(e, e > 2)   // [3, 4]
```

### fold

Combines all elements of a collection using a binary function.

```
// For lists:
[1, 2, 3].fold(e, acc, acc + e)   // 6

// For maps:
{"a": "apple", "b": "banana"}.fold(k, v, acc, acc + v)   // "applebanana"

// Build a map from a list of key/value objects:
dyn(tags).fold(tag, acc, merge(acc, {tag.key: tag.value}))
```

When folding a variable declared as `any`, wrap it with `dyn(...)` so CEL can use it as a comprehension range. The `merge(left, right)` helper returns a map containing all keys from both maps; values from `right` replace values from `left` on duplicate keys.

### has

Tests whether a field is available in a message or map.

```
has(person.name)    // true if 'name' is present, false otherwise
```

### in

Membership test operator.

```
"apple" in ["apple", "banana"]    // true
3 in [1, 2, 4]                    // false
```

### map

Creates a new list by transforming each element.

```
[1, 2, 3].map(e, e * 2)           // [2, 4, 6]
[1, 2, 3].map(x, x > 1, x + 1)   // [3, 4]
```

### or

If the left-hand side is none-type, return the right-hand side optional value.

```
obj.?field.or(m[?key])
l[?index].or(obj.?field.subfield).or(obj.?other)
```

### orValue

Returns the value if present, otherwise returns a default.

```
{'a': 'x', 'b': 'y', 'c': 'z'}.?c.orValue('empty')   // "z"
{'a': 'x', 'b': 'y'}.?c.orValue('empty')              // "empty"
[1, 2, 3][?2].orValue(5)                               // 3
[1, 2][?2].orValue(5)                                  // 5
```

### size

Returns the number of elements in a collection or characters in a string.

```
"apple".size()                      // 5
b"abc".size()                       // 3
["apple", "banana", "cherry"].size() // 3
{"a": 1, "b": 2}.size()             // 2
```

### slice

Returns a new sub-list using the given indices.

```
[1, 2, 3, 4].slice(1, 3)    // [2, 3]
[1, 2, 3, 4].slice(2, 4)    // [3, 4]
```

---

## sets

### sets.contains

Returns whether the first list contains all elements of the second list.

```
sets.contains([], [])               // true
sets.contains([], [1])              // false
sets.contains([1, 2, 3, 4], [2, 3]) // true
```

### sets.equivalent

Returns whether two lists are set-equivalent.

```
sets.equivalent([], [])             // true
sets.equivalent([1], [1, 1])        // true
sets.equivalent([1, 2, 3], [3u, 2.0, 1]) // true
```

### sets.intersects

Returns whether the two lists share at least one common element.

```
sets.intersects([1], [])            // false
sets.intersects([1], [1, 2])        // true
```

---

## csv

### CSV

Converts a CSV formatted array into a two-dimensional array.

```
CSV(["Alice,30", "Bob,31"])[0][0]   // "Alice"
```

---

## crypto

### crypto.SHA1 | SHA256 | SHA384 | SHA512

Computes a SHA hash of the input data.

```
crypto.SHA1("hello")    // "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"
crypto.SHA256("hello")  // "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
```

---

## dates

### timestamp

Represents a point in time.

```
timestamp("2023-01-01T00:00:00Z")
timestamp("2023-07-04T12:00:00Z")
```

### .getDate

Extracts the date part from a timestamp.

```
"2023-01-01T12:34:56Z".getDate()    // "2023-01-01"
```

### Date Part Accessors

| Function | Description | Range |
|---|---|---|
| `<date>.getDayOfMonth()` | Day of the month | 1–31 |
| `<date>.getDayOfWeek()` | Day of the week (Sunday=0) | 0–6 |
| `<date>.getDayOfYear()` | Day of the year | 1–366 |
| `<date>.getFullYear()` | Full 4-digit year | |
| `<date>.getHours()` | Hour | 0–23 |
| `<date>.getMilliseconds()` | Milliseconds | 0–999 |
| `<date>.getMinutes()` | Minutes | 0–59 |
| `<date>.getMonth()` | Month (January=0) | 0–11 |
| `<date>.getSeconds()` | Seconds | 0–59 |

### duration

Parses a string into a duration.

```
duration("5h")    // 5 hours
duration("30m")   // 30 minutes
duration("7d")    // 7 days
```

### time.ZoneName

Returns the name of the local system's time zone.

```
time.ZoneName()   // "PST"
```

### time.ZoneOffset

Returns the offset of the local system's time zone in minutes.

```
time.ZoneOffset()   // -480 (for PST)
```

### time.Parse

Parses a string into a time object using a specified layout.

```
time.Parse("2006-01-02", "2023-09-26")
time.Parse("02-01-2006", "26-09-2023")
time.Parse("15:04 02-01-2006", "14:30 26-09-2023")
```

### time.ParseLocal

Parses a string into a time object using the local time zone.

```
time.ParseLocal("2006-01-02 15:04", "2023-09-26 14:30")
```

### time.ParseInLocation

Parses a string into a time object for a specific time zone.

```
time.ParseInLocation("2006-01-02", "America/New_York", "2023-09-26")
time.ParseInLocation("02-01-2006", "Europe/London", "26-09-2023")
time.ParseInLocation("15:04 02-01-2006", "Asia/Tokyo", "14:30 26-09-2023")
```

### time.Now

Returns the current time.

```
time.Now()
```

### time.ParseDuration

Parses a string into a duration with support for days.

```
time.ParseDuration("1h30m")    // 1 hour 30 minutes
time.ParseDuration("7d")       // 7 days
time.ParseDuration("30d12h")   // 30 days and 12 hours
time.ParseDuration("-2h45m")   // negative duration
```

### time.Since

Calculates the duration elapsed since a given time.

```
time.Since(time.Parse("2006-01-02", "2023-09-26"))
time.Since(time.Now())    // very small duration
```

### time.Until

Calculates the duration remaining until a specified future time.

```
time.Until(time.Parse("2006-01-02", "2023-10-01"))
```

---

## encode

### urlencode

Encodes a string as URL-encoded.

```
urlencode("hello world ?")    // "hello+world+%3F"
```

### urldecode

Decodes a URL-encoded string.

```
urldecode("hello+world+%3F")  // "hello world ?"
```

---

## filepath

### filepath.Base

Returns the last element of path.

```
filepath.Base("/home/user/projects/gencel")   // "gencel"
```

### filepath.Clean

Returns the shortest path equivalent by lexical processing.

```
filepath.Clean("/foo/bar/../baz")    // "/foo/baz"
```

### filepath.Dir

Returns all but the last element of path (the directory).

```
filepath.Dir("/home/user/projects/gencel")   // "/home/user/projects"
```

### filepath.Ext

Returns the file name extension.

```
filepath.Ext("/opt/image.jpg")    // ".jpg"
```

### filepath.IsAbs

Reports whether the path is absolute.

```
filepath.IsAbs("/home/user/projects/gencel")   // true
filepath.IsAbs("projects/gencel")              // false
```

### filepath.Join

Joins path elements into a single path.

```
filepath.Join(["/home/user", "projects", "gencel"])   // "/home/user/projects/gencel"
```

### filepath.Match

Reports whether a name matches the shell file name pattern.

```
filepath.Match("*.txt", "foo.json")   // false
filepath.Match("*.txt", "foo.txt")    // true
```

### filepath.Rel

Returns a relative path from basepath to targpath.

```
filepath.Rel("/foo/bar", "/foo/bar/baz")   // "baz"
```

### filepath.Split

Splits path into directory and file name components.

```
filepath.Split("/foo/bar/baz")   // ["/foo/bar/" "baz"]
```

---

## JSON

### .JSON

Parses a string into an object.

```
'{"name": "Alice", "age": 30}'.JSON()
```

### .JSONArray

Parses a string into an array.

```
'[{"name": "Alice"}, {"name": "Bob"}]'.JSONArray()
```

### .toJSON

Converts an object into a JSON formatted string.

```
[{ name: "John" }].toJSON()      // '[{"name":"John"}]'
{'name': 'John'}.toJSON()        // '{"name":"John"}'
1.toJSON()                       // "1"
```

### .toJSONPretty

Converts data into a JSON string with indentation.

```
{'name': 'aditya'}.toJSONPretty('\t')
// {
//   "name": "aditya"
// }
```

### jmespath

Evaluates a [JMESPath](https://jmespath.org/) expression against an object.

```
jmespath("city", { name: "John", age: 30, city: "NY" })   // "NY"
```

### jsonpath

Evaluates a [JSONPath](https://datatracker.ietf.org/doc/draft-ietf-jsonpath-base) expression against an object.

```
jsonpath("$.name", { name: "John", age: 30 })                 // "John"
jsonpath("$.items[0]", { items: ["apple", "banana"] })         // "apple"
jsonpath("$.addresses[-1:].city", { addresses: [{city:"NYC"},{city:"SF"}] }) // "SF"
jsonpath("$.user.email", '{"user": {"email": "john@example.com"}}')  // "john@example.com"
```

### jq

Applies a jq expression to filter or transform data.

```
jq(".name", { name: "John", age: 30 })                 // "John"
jq("{name, age}", { name: "John", age: 30, city: "NY" }) // {"name":"John","age":30}
jq(".[] | select(.age > 25)", [{ name: "John", age: 30 }, { name: "Jane", age: 25 }])
// [{"name": "John", "age": 30}]
```

---

## kubernetes

### k8s.cpuAsMillicores

Returns the millicores of a Kubernetes CPU resource string.

```
k8s.cpuAsMillicores("10m")    // 10
k8s.cpuAsMillicores("0.5")    // 500
k8s.cpuAsMillicores("1.234")  // 1234
```

### k8s.getHealth

Retrieves the health status of a Kubernetes resource as a map.

```
k8s.getHealth(pod)        // map with health info
k8s.getHealth(service)
k8s.getHealth(deployment)
```

### k8s.getStatus

Retrieves the status of a Kubernetes resource as a string.

```
k8s.getStatus(pod)         // "Running"
k8s.getStatus(service)     // "Active"
k8s.getStatus(deployment)  // "Deployed"
```

### k8s.getResourcesLimit

Retrieves the CPU or memory limit of a Kubernetes pod.

```
k8s.getResourcesLimit(pod, "cpu")    // 2
k8s.getResourcesLimit(pod, "memory") // 200
```

### k8s.getResourcesRequests

Retrieves the CPU or memory requests of a Kubernetes pod.

```
k8s.getResourcesRequests(pod, "cpu")    // 2
k8s.getResourcesRequests(pod, "memory") // 200
```

### k8s.isHealthy

Determines if a Kubernetes resource is healthy.

```
k8s.isHealthy(pod)        // true
k8s.isHealthy(service)    // false
k8s.isHealthy(deployment) // true
```

### k8s.labels

Returns a map of all labels for a Kubernetes object. The namespace is included for namespaced objects, and labels ending with `-hash` are excluded.

```
k8s.labels(pod)   // {"namespace": "kube-system", "app": "kube-dns"}
```

### k8s.memoryAsBytes

Converts a memory string to bytes.

```
k8s.memoryAsBytes("10Ki")      // 10240
k8s.memoryAsBytes("1.234Gi")   // 1324997410
```

### k8s.nodeProperties

Returns a map of node properties (CPU, memory, ephemeral-storage, zone).

```
k8s.nodeProperties(node)
// {"cpu": 2, "memory": 200, "ephemeral-storage": 100000, "zone": "us-east-1a"}
```

### k8s.podProperties

Returns a map of pod properties (image, cpu, memory, node, created-at, namespace).

```
k8s.podProperties(pod)
// {"image": "postgres:14", "node": "saka", ...}
```

---

## math

### math.Add

Sums a list of numbers.

```
math.Add([1, 2, 3, 4, 5])   // 15
```

### math.Sub

Subtracts the second number from the first.

```
math.Sub(5, 4)   // 1
```

### math.Mul

Returns the product of a list of numbers.

```
math.Mul([1, 2, 3, 4, 5])   // 120
```

### math.Div

Divides the first number by the second.

```
math.Div(4, 2)   // 2
```

### math.Rem

Returns the remainder of dividing the first number by the second.

```
math.Rem(4, 3)   // 1
```

### math.Pow

Returns the result of raising the first number to the power of the second.

```
math.Pow(4, 2)   // 16
```

### math.Seq

Generates a sequence of numbers.

```
math.Seq([1, 5])        // [1, 2, 3, 4, 5]
math.Seq([1, 6, 2])     // [1, 3, 5]
```

### math.Abs

Returns the absolute value of a number.

```
math.Abs(-1)    // 1
```

### math.greatest

Returns the greatest value in a list.

```
math.greatest([1, 2, 3, 4, 5])   // 5
```

### math.least

Returns the least value in a list.

```
math.least([1, 2, 3, 4, 5])   // 1
```

### math.Ceil

Returns the smallest integer greater than or equal to the input.

```
math.Ceil(2.3)   // 3
```

### math.Floor

Returns the largest integer less than or equal to the input.

```
math.Floor(2.3)   // 2
```

### math.Round

Returns the nearest integer to the input.

```
math.Round(2.3)   // 2
math.Round(2.7)   // 3
math.Round(2.5)   // 3
```

### math.Trunc

Returns the integer part of a float, truncating towards zero.

```
math.Trunc(2.7)    // 2
math.Trunc(-2.7)   // -2
```

### math.Sign

Returns the sign of a number: -1, 0, or 1.

```
math.Sign(-5)   // -1
math.Sign(0)    // 0
math.Sign(5)    // 1
```

### math.Sqrt

Returns the square root of a number.

```
math.Sqrt(16)   // 4
math.Sqrt(2)    // 1.4142135623730951
```

### math.IsNaN

Checks if a value is Not-a-Number.

```
math.IsNaN(0.0 / 0.0)   // true
math.IsNaN(5.0)         // false
```

### math.IsInf

Checks if a value is infinite.

```
math.IsInf(1.0 / 0.0)   // true
math.IsInf(5.0)         // false
```

### math.IsFinite

Checks if a value is finite.

```
math.IsFinite(5.0)         // true
math.IsFinite(1.0 / 0.0)   // false
```

### Bitwise Operations

| Function | Description | Example |
|---|---|---|
| `math.bitOr(x, y)` | Bitwise OR | `math.bitOr(5, 3)` → `7` |
| `math.bitAnd(x, y)` | Bitwise AND | `math.bitAnd(5, 3)` → `1` |
| `math.bitXor(x, y)` | Bitwise XOR | `math.bitXor(5, 3)` → `6` |
| `math.bitNot(x)` | Bitwise NOT | `math.bitNot(5)` → `-6` |
| `math.bitShiftLeft(x, n)` | Left shift | `math.bitShiftLeft(5, 1)` → `10` |
| `math.bitShiftRight(x, n)` | Right shift | `math.bitShiftRight(10, 1)` → `5` |

---

## random

### random.ASCII

Generates a random ASCII string of a specified length.

```
random.ASCII(5)
```

### random.Alpha

Generates a random alphabetic string of a specified length.

```
random.Alpha(5)
```

### random.AlphaNum

Generates a random alphanumeric string of a specified length.

```
random.AlphaNum(5)
```

### random.String

Generates a random string of a specified length and optional character set.

```
random.String(5)
random.String(5, ["a", "d"])   // 5 chars between 'a' and 'd'
```

### random.Item

Returns a random item from a list.

```
random.Item(["a", "b", "c"])
```

### random.Number

Returns a random integer within a specified range.

```
random.Number(1, 10)
```

### random.Float

Returns a random float within a specified range.

```
random.Float(1, 10)
```

---

## regexp

### regexp.Find

Finds the first occurrence of a pattern within a string.

```
regexp.Find("llo", "hello")         // "llo"
regexp.Find("\\d+", "abc123def")    // "123"
regexp.Find("xyz", "hello")         // ""
```

### regexp.FindAll

Retrieves all occurrences of a pattern within a string.

```
regexp.FindAll("a.", -1, "banana")    // ["ba", "na", "na"]
regexp.FindAll("\\d", 2, "12345")     // ["1", "2"]
regexp.FindAll("z", -1, "hello")      // []
```

### regexp.Match

Checks if a string matches a regular expression pattern.

```
regexp.Match("^h.llo", "hello")    // true
regexp.Match("^b", "apple")        // false
regexp.Match("\\d+", "abc123")     // true
```

### regexp.QuoteMeta

Quotes all regular expression metacharacters in a string.

```
regexp.QuoteMeta("a.b")     // "a\\.b"
regexp.QuoteMeta("abc")     // "abc"
```

### regexp.Replace

Replaces occurrences of a pattern within a string.

```
regexp.Replace("a.", "x", "banana")      // "bxnxna"
regexp.Replace("z", "x", "apple")        // "apple"
regexp.Replace("\\d+", "num", "abc123")  // "abcnum"
```

### regexp.ReplaceLiteral

Replaces occurrences of a substring without regex interpretation.

```
regexp.ReplaceLiteral("apple", "orange", "apple pie")   // "orange pie"
regexp.ReplaceLiteral("a.", "x", "a.b c.d")            // "x b c.d"
```

### regexp.Split

Splits a string into substrings separated by a pattern.

```
regexp.Split("a.", -1, "banana")           // ["", "n", "n"]
regexp.Split("\\s", 2, "apple pie is delicious")  // ["apple", "pie is delicious"]
regexp.Split("z", -1, "hello")             // ["hello"]
```

---

## strings

### .abbrev

Abbreviates a string using ellipses.

```
"Now is the time for all good men".abbrev(5, 20)   // "...s the time for..."
"KubernetesPod".abbrev(1, 5)                       // "Ku..."
"KubernetesPod".abbrev(6)                          // "Kub..."
```

### .camelCase

Converts a string to camelCase format.

```
"hello world".camelCase()               // "HelloWorld"
"hello_world".camelCase()               // "HelloWorld"
"hello beautiful world!".camelCase()    // "HelloBeautifulWorld"
```

### .charAt

Returns the character at the given position.

```
"hello".charAt(4)    // "o"
```

### .contains

Checks if a string contains a given substring.

```
"apple".contains("app")   // true
```

### .endsWith

Determines if a string ends with a specified substring.

```
"hello".endsWith("lo")   // true
```

### .format

Creates a new string with printf-style substitutions.

```
"this is a string: %s\nand an integer: %d".format(["str", 42])
// "this is a string: str\nand an integer: 42"
```

### .indent

Indents each line of a string by the specified width and prefix.

```
"hello world".indent(4, "-")   // "----hello world"
```

### .indexOf

Returns the integer index of the first occurrence of a substring.

```
"hello mellow".indexOf("")        // 0
"hello mellow".indexOf("ello")    // 1
"hello mellow".indexOf("jello")   // -1
"hello mellow".indexOf("", 2)     // 2
```

### .join

Concatenates elements of a string list.

```
["hello", "mellow"].join()      // "hellomellow"
["hello", "mellow"].join(" ")   // "hello mellow"
[].join("/")                    // ""
```

### .kebabCase

Converts a string to kebab-case format.

```
"Hello World".kebabCase()            // "hello-world"
"HelloWorld".kebabCase()             // "hello-world"
"Hello Beautiful World!".kebabCase() // "hello-beautiful-world"
```

### .lastIndexOf

Returns the index of the last occurrence of a substring.

```
"hello mellow".lastIndexOf("")        // 12
"hello mellow".lastIndexOf("ello")    // 7
"hello mellow".lastIndexOf("jello")   // -1
"hello mellow".lastIndexOf("ello", 6) // 1
```

### .lowerAscii

Returns a new string with all ASCII characters lower-cased.

```
"TacoCat".lowerAscii()       // "tacocat"
"TacoCÆt Xii".lowerAscii()   // "tacocÆt xii"
```

### .matches

Determines if a string matches a regular expression pattern.

```
"apple".matches("^a.*e$")                                            // true
"example@email.com".matches("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$") // true
"12345".matches("^\\d+$")                                            // true
```

Built-in CEL regex operations, including `.matches()`, reject regex programs larger than 10,000 instructions. This limit applies to both literal and dynamically supplied patterns. It does not apply to gomplate's separate `regexp.*` functions.

### .quote

Makes a string safe to print by escaping special characters.

```
strings.quote('single-quote with "double quote"')
// '"single-quote with \"double quote\""'
```

### .repeat

Repeats a string a given number of times.

```
"apple".repeat(3)   // "appleappleapple"
```

### .replace

Replaces occurrences of a substring with a replacement string.

```
"hello hello".replace("he", "we")       // "wello wello"
"hello hello".replace("he", "we", 1)    // "wello hello"
"hello hello".replace("he", "we", 0)    // "hello hello"
```

### .replaceAll

Replaces all occurrences of a substring.

```
"I have an apple".replaceAll("apple", "orange")   // "I have an orange"
```

### .reverse

Returns a string with characters in reverse order.

```
"gums".reverse()          // "smug"
"John Smith".reverse()    // "htimS nhoJ"
```

### .runeCount

Counts the number of runes in a string.

```
"Hello World".runeCount()   // 11
```

### .shellQuote

Quotes a string for safe use as a shell token.

```
"Hello World".shellQuote()        // "'Hello World'"
"Hello$World".shellQuote()        // "'Hello$World'"
```

### .size

Returns the number of characters in a string or elements in a collection.

```
["apple", "banana", "cherry"].size()   // 3
"hello".size()                         // 5
```

### .slug

Converts a string into a URL-friendly slug.

```
"Hello World!".slug()           // "hello-world"
"Hello, World!".slug()          // "hello-world"
"Hello Beautiful World".slug()  // "hello-beautiful-world"
```

### .snakeCase

Converts a string to snake_case format.

```
"Hello World".snakeCase()            // "hello_world"
"HelloWorld".snakeCase()             // "hello_world"
"Hello Beautiful World!".snakeCase() // "hello_beautiful_world"
```

### .sort

Sorts a string alphabetically.

```
"hello".sort()   // "ehllo"
```

### .split

Splits a string by a separator.

```
"hello hello hello".split(" ")       // ["hello", "hello", "hello"]
"hello hello hello".split(" ", 2)    // ["hello", "hello hello"]
"hello hello hello".split(" ", -1)   // ["hello", "hello", "hello"]
```

### .squote

Adds single quotes around a string.

```
"Hello World".squote()   // "'Hello World'"
```

### .startsWith

Determines if a string starts with a specified substring.

```
"hello".startsWith("he")   // true
```

### .substring

Returns a substring given a numeric range.

```
"tacocat".substring(4)     // "cat"
"tacocat".substring(0, 4)  // "taco"
```

### .title

Converts the first character of each word to uppercase.

```
"hello world".title()   // "Hello World"
```

### .trim

Removes leading and trailing whitespace.

```
"  \ttrim\n    ".trim()   // "trim"
```

### .trimPrefix

Removes a prefix from a string.

```
"hello world".trimPrefix("hello ")   // "world"
```

### .trimSuffix

Removes a suffix from a string.

```
"hello world".trimSuffix(" world")   // "hello"
```

### .upperAscii

Returns a new string with all ASCII characters upper-cased.

```
"TacoCat".upperAscii()   // "TACOCAT"
```
