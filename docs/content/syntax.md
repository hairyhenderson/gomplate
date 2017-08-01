---
title: Syntax
weight: 12
menu: main
---

## About `.Env`

You can easily access environment variables with `.Env`, but there's a catch:
if you try to reference an environment variable that doesn't exist, parsing
will fail and `gomplate` will exit with an error condition.

Sometimes, this behaviour is desired; if the output is unusable without certain
strings, this is a sure way to know that variables are missing!

If you want different behaviour, try [`getenv`](../functions/env/#env-getenv).

## Built-in functions

In addition to all of the functions and operators that the [Go template](https://golang.org/pkg/text/template/)
language provides (`if`, `else`, `eq`, `and`, `or`, `range`, etc...), there are
some additional functions baked in to `gomplate`. See the links on the left.