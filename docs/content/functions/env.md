---
title: env functions
menu:
  main:
    parent: functions
---

## `env.Getenv`

**Alias:** `getenv`

Exposes the [os.Getenv](https://golang.org/pkg/os/#Getenv) function.

This is a more forgiving alternative to using `.Env`, since missing keys will
return an empty string.

An optional default value can be given as well.

#### Example

```console
$ gomplate -i 'Hello, {{env.Getenv "USER"}}'
Hello, hairyhenderson
$ gomplate -i 'Hey, {{getenv "FIRSTNAME" "you"}}!'
Hey, you!
```