---
title: env functions
menu:
  main:
    parent: functions
---

## `env.Getenv`

**Alias:** `getenv`

Exposes the [os.Getenv](https://golang.org/pkg/os/#Getenv) function.

Retrieves the value of the environment variable named by the key. If the 
variable is unset, but the same variable ending in `_FILE` is set, the contents
of the file will be returned. Otherwise the provided default (or an empty
string) is returned.

This is a more forgiving alternative to using `.Env`, since missing keys will
return an empty string, instead of panicking.

The `_FILE` fallback is especially useful for use with [12-factor][]-style
applications configurable only by environment variables, and especially in
conjunction with features like [Docker Secrets][].

#### Example

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

## `env.ExpandEnv`

Exposes the [os.ExpandEnv](https://golang.org/pkg/os/#ExpandEnv) function.

Replaces `${var}` or `$var` in the input string according to the values of the
current environment variables. References to undefined variables are replaced by the empty string.

Like [`env.Getenv`](#env-getenv), the `_FILE` variant of a variable is used.

#### Example

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

[12-factor]: https://12factor.net
[Docker Secrets]: https://docs.docker.com/engine/swarm/secrets/#build-support-for-docker-secrets-into-your-images
