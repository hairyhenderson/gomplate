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

[12-factor]: https://12factor.net
[Docker Secrets]: https://docs.docker.com/engine/swarm/secrets/#build-support-for-docker-secrets-into-your-images