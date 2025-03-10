---
title: semver functions
menu:
  main:
    parent: functions
---

These functions allow user you to parse a [semantic version](http://semver.org/) string or test it with constraint.

It's implemented with the https://github.com/Masterminds/semver library.

## `semver.Semver`_(unreleased)_
**Unreleased:** _This function is in development, and not yet available in released builds of gomplate._

Returns a semantic version struct holding the `input` version string.

The returned struct are defined at: [`semver.Version`](https://pkg.go.dev/github.com/Masterminds/semver/v3#Version).

### Usage

```
semver.Semver input
```
```
input | semver.Semver
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ The input to parse |

### Examples

```console
$ gomplate -i '{{ semver.Semver "v1.1.1"}}'
1.1.1
```
```console
$ gomplate -i '{{ (semver.Semver "v1.1.1").Major }}'
1
```
```console
$ gomplate -i 'the pre release version is {{ ("v1.1.1" | semver.Semver).SetPrerelease "beta.1" }}'
the pre release version is 1.1.1-beta.1
```

## `semver.CheckConstraint`_(unreleased)_
**Unreleased:** _This function is in development, and not yet available in released builds of gomplate._

Test whether the input version matches the constraint.

Ref: https://github.com/Masterminds/semver#checking-version-constraints

### Usage

```
semver.CheckConstraint constraint input
```
```
input | semver.CheckConstraint constraint
```

### Arguments

| name | description |
|------|-------------|
| `constraint` | _(required)_ The constraints expression to test. |
| `input` | _(required)_ The input semantic version string to test. |

### Examples

```console
$ gomplate -i '{{ semver.CheckConstraint "> 1.0" "v1.1.1" }}'
true
```
```console
$ gomplate -i '{{ semver.CheckConstraint "> 1.0, <1.1" "v1.1.1" }}'
false
```
```console
$ gomplate -i '{{ "v1.1.1" | semver.CheckConstraint "> 1.0" }}'
true
```
