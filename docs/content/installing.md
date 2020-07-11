---
title: Installing
weight: 10
menu: main
---

## macOS with homebrew

The simplest method for macOS is to use homebrew:

```console
$ brew install gomplate
...
```

## Alpine Linux

`gomplate` is available in Alpine's `community` repository.

```console
$ apk add --no-cache gomplate
...
```

_Note: the Alpine version of gomplate may lag behind the latest release of gomplate._

## use with Docker

A simple way to get started is with one of the [hairyhenderson/gomplate][] Docker images. Images containing [`slim` binaries](#slim-binaries) are tagged as `:slim` or `:vX.Y.Z-slim`.

```console
$ docker run hairyhenderson/gomplate --version
```

Of course, there are some drawbacks - any files to be used for [datasources][]
must be mounted and any environment variables to be used must be passed through:

```console
$ echo 'My voice is my {{.Env.THING}}. {{(datasource "vault").value}}' \
  | docker run -i -e THING=passport -v /home/me/.vault-token:/root/.vault-token hairyhenderson/gomplate -d vault=vault:///secret/sneakers -f -
My voice is my passport. Verify me.
```

It can be pretty awkward to always type `docker run hairyhenderson/gomplate`,
so this can be made simpler with a shell alias:

```console
$ alias gomplate=docker run hairyhenderson/gomplate
$ gomplate --version
gomplate version 1.2.3
```

### use inside a container

`gomplate` is often used inside Docker containers. When building images with Docker 17.05 or higher, you can use [multi-stage builds][] to easily include the `gomplate` binary in your container images.

Use the `COPY` instruction's `--from` flag to accomplish this:

```Dockerfile
...
COPY --from=hairyhenderson/gomplate:v2.5.0-slim /gomplate /bin/gomplate
```

Now, `gomplate` will be available in the `/bin` directory inside the container image.

Note that when using `gomplate` with HTTPS-based datasources, you will likely need to install the `ca-certificates` package for your base distribution. Here's an example when using the [`alpine`](https://hub.docker.com/_alpine) base image:

```Dockerfile
FROM alpine

COPY --from=hairyhenderson/gomplate:v2.5.0-slim /gomplate /bin/gomplate
RUN apk add --no-cache ca-certificates
```

## manual install

1. Get the latest `gomplate` for your platform from the [releases][] page
    - if available, you may want to download the [`-slim` variant](#slim-binaries)
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
$ go get github.com/hairyhenderson/gomplate/v3/cmd/gomplate
$ gomplate --help
...
```

## install with `npm`

For some users, especially Node.js developers, using `npm` may be a natural fit.
Even though `gomplate` is written in Go and not Node.js, it can still be installed
with `npm`:

```console
$ npm install -g gomplate
...
```

## Slim binaries

As a convenience, self-extracting compressed `gomplate` binaries are available from the [releases][] page. These are named with `-slim` as a suffix (or `-slim.exe`). They are compressed with [UPX][].

Generally, these binaries are ~5x smaller than the regular ones, but are otherwise exactly the same.

There are a few reasons that a regular binary is also distributed:
- UPX lacks support for some platforms
- there's a very slight chance that the slim binary could exhibit some form of bug related to being compressed
- there could be environments where self-extracting compressed executables are disallowed

[releases]: https://github.com/hairyhenderson/gomplate/releases
[UPX]: https://upx.github.io/
[multi-stage builds]: https://docs.docker.com/develop/develop-images/multistage-build/
[hairyhenderson/gomplate]: https://hub.docker.com/r/hairyhenderson/gomplate/tags/
