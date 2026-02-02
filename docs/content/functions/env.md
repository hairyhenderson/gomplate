---
title: env functions
menu:
  main:
    parent: functions
---

[12-factor]: https://12factor.net
[Docker Secrets]: https://docs.docker.com/engine/swarm/secrets/#build-support-for-docker-secrets-into-your-images

## `env.Getenv`

**Alias:** `getenv`

Exposes the [os.Getenv](https://pkg.go.dev/os/#Getenv) function.

Retrieves the value of the environment variable named by the key. If the
variable is unset, but the same variable ending in `_FILE` is set, the contents
of the file will be returned. Otherwise the provided default (or an empty
string) is returned.

This is a more forgiving alternative to using `.Env`, since missing keys will
return an empty string, instead of panicking.

The `_FILE` fallback is especially useful for use with [12-factor][]-style
applications configurable only by environment variables, and especially in
conjunction with features like [Docker Secrets][].

_Added in gomplate [v0.2.0](https://github.com/hairyhenderson/gomplate/releases/tag/v0.2.0)_
### Usage

```
env.Getenv var [default]
```

### Arguments

| name | description |
|------|-------------|
| `var` | _(required)_ the environment variable name |
| `default` | _(optional)_ the default |

### Examples

```console
$ gomplate -i 'Hello, {{env.Getenv "USER"}}'
Hello, hairyhenderson
$ gomplate -i 'Hey, {{getenv "FIRSTNAME" "you"}}!'
Hey, you!
```
```console
$ echo "safe" > /tmp/mysecret
$ export SECRET_FILE=/tmp/mysecret
$ gomplate -i 'Your secret is {{getenv "SECRET"}}'
Your secret is safe
```

## `env.Env`

Returns a map of all environment variables. This is equivalent to [`.Env`](/syntax/#env),
but accessible via the `env` namespace.

Unlike [`env.Getenv`](#envgetenv), this does not read `_FILE` variants, and will
fail when accessing a missing key.

This is useful when you want strict environment variable access without the
`_FILE` fallback behavior.

_Added in gomplate [v5.1.0](https://github.com/hairyhenderson/gomplate/releases/tag/v5.1.0)_
### Usage

```
env.Env
```


### Examples

```console
$ gomplate -i 'Hello, {{env.Env.USER}}'
Hello, hairyhenderson
```
```console
$ gomplate -i '{{ env.Env.HOME }}'
/home/hairyhenderson
```

## `env.HasEnv`

Returns `true` if the environment variable is set, `false` otherwise.
This wraps the [`os.LookupEnv`](https://pkg.go.dev/os/#LookupEnv) function.

Note that a variable set to an empty string is still considered "set".

_Added in gomplate [v5.1.0](https://github.com/hairyhenderson/gomplate/releases/tag/v5.1.0)_
### Usage

```
env.HasEnv var
```
```
var | env.HasEnv
```

### Arguments

| name | description |
|------|-------------|
| `var` | _(required)_ the environment variable name |

### Examples

```console
$ gomplate -i '{{if env.HasEnv "FOO"}}FOO is set{{else}}FOO is not set{{end}}'
FOO is not set
$ FOO=bar gomplate -i '{{if env.HasEnv "FOO"}}FOO is set{{else}}FOO is not set{{end}}'
FOO is set
```
```console
$ EMPTY= gomplate -i '{{if env.HasEnv "EMPTY"}}EMPTY is set{{end}}'
EMPTY is set
```
```console
$ gomplate -i '{{ "USER" | env.HasEnv }}'
true
```

## `env.ExpandEnv`

Exposes the [os.ExpandEnv](https://pkg.go.dev/os/#ExpandEnv) function.

Replaces `${var}` or `$var` in the input string according to the values of the
current environment variables. References to undefined variables are replaced by the empty string.

Like [`env.Getenv`](#envgetenv), the `_FILE` variant of a variable is used.

_Added in gomplate [v2.5.0](https://github.com/hairyhenderson/gomplate/releases/tag/v2.5.0)_
### Usage

```
env.ExpandEnv input
```

### Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input |

### Examples

```console
$ gomplate -i '{{env.ExpandEnv "Hello $USER"}}'
Hello, hairyhenderson
$ gomplate -i 'Hey, {{env.ExpandEnv "Hey, ${FIRSTNAME}!"}}'
Hey, you!
```
```console
$ echo "safe" > /tmp/mysecret
$ export SECRET_FILE=/tmp/mysecret
$ gomplate -i '{{env.ExpandEnv "Your secret is $SECRET"}}'
Your secret is safe
```
```console
$ gomplate -i '{{env.ExpandEnv (file.Read "foo")}}
contents of file "foo"...
```
