---
title: Installing
weight: 10
---
# Installing

## macOS with homebrew

The simplest method for macOS is to use homebrew:

```console
$ brew install hairyhenderson/tap/gomplate
...
```

## Alpine Linux

Currently, `gomplate` is available in the `community` repository for the `edge` release.

```console
$ echo "http://dl-cdn.alpinelinux.org/alpine/edge/community/" >> /etc/apk/repositories
$ apk update
$ apk add gomplate
...
```

_Note: the Alpine version of gomplate may lag behind the latest release of gomplate._

## use with Docker

A simple way to get started is with the Docker image.

```console
$ docker run hairyhenderson/gomplate --version
```

Of course, there are some drawbacks - any files to be used for [datasources][]
must be mounted and any environment variables to be used must be passed through:

```console
$ echo 'My voice is my {{.Env.THING}}. {{(datasource "vault").value}}' \
  | docker run -e THING=passport -v /home/me/.vault-token:/root/.vault-token hairyhenderson/gomplate -d vault=vault:///secret/sneakers
My voice is my passport. Verify me.
```

It can be pretty awkward to always type `docker run hairyhenderson/gomplate`,
so this can be made simpler with a shell alias:

```console
$ alias gomplate=docker run hairyhenderson/gomplate
$ gomplate --version
gomplate version 1.2.3
```

## manual install

1. Get the latest `gomplate` for your platform from the [releases](https://github.com/hairyhenderson/gomplate/releases) page
2. Store the downloaded binary somewhere in your path as `gomplate` (or `gomplate.exe`
  on Windows)
3. Make sure it's executable (on Linux/macOS)
3. Test it out with `gomplate --help`!

In other words:

```console
$ curl -o /usr/local/bin/gomplate -sSL https://github.com/hairyhenderson/gomplate/releases/download/<version>/gomplate_<os>-<arch>
$ chmod 755 /usr/local/bin/gomplate
$ gomplate --help
...
```

## install with `go get`

If you're a Go user already, sometimes it's faster to just use `go get` to install `gomplate`:

```console
$ go get github.com/hairyhenderson/gomplate
$ gomplate --help
...
```
