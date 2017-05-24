---
title: Syntax
weight: 0
---

## About `.Env`

You can easily access environment variables with `.Env`, but there's a catch:
if you try to reference an environment variable that doesn't exist, parsing
will fail and `gomplate` will exit with an error condition.

Sometimes, this behaviour is desired; if the output is unusable without certain
strings, this is a sure way to know that variables are missing!

If you want different behaviour, try [`getenv`](../functions/#getenv).
