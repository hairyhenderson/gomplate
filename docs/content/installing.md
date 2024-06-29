---
title: Installing
weight: 10
menu: main
---

There are many installation methods available for gomplate, depending on your platform and use-case.

## macOS/Linux with homebrew

The simplest method for macOS and Linux is to use [homebrew](https://brew.sh/):

```console
$ brew install gomplate
...

==> Installing gomplate
==> Pouring gomplate-3.8.0.x86_64_linux.bottle.tar.gz
🍺  /home/linuxbrew/.linuxbrew/Cellar/gomplate/3.8.0: 6 files, 7.8MB
```

## macOS with MacPorts

On macOS, you can also install gomplate using [MacPorts](https://www.macports.org):

```console
$ sudo port install gomplate
```

## Windows with Chocolatey

The simplest method for installing gomplate on Windows is to use [`choco`](https://community.chocolatey.org/packages/gomplate):

```console
choco install gomplate
```

## Alpine Linux

`gomplate` is available in Alpine's `community` repository.

```console
$ apk add --no-cache gomplate
...
```

_Note: the Alpine version of gomplate may lag behind the latest release of gomplate._

## use with Docker

A simple way to get started is with one of the [hairyhenderson/gomplate][] Docker images.

```console
$ docker run hairyhenderson/gomplate:stable --version
gomplate version 3.9.0
```

Of course, there are some drawbacks - any files to be used for [datasources][]
must be mounted and any environment variables to be used must be passed through:

```console
$ echo 'My voice is my {{.Env.THING}}. {{(datasource "vault").value}}' | docker run -i -e THING=passport -v /home/me/.vault-token:/root/.vault-token hairyhenderson/gomplate -d vault=vault:///secret/sneakers -f -
My voice is my passport. Verify me.
```

It can be awkward to always type `docker run hairyhenderson/gomplate:stable`,
so this can be made simpler with a shell alias:

```console
$ alias gomplate='docker run hairyhenderson/gomplate:stable'
$ gomplate --version
gomplate version 3.9.0
```

### use inside a container

`gomplate` is often used inside Docker containers. When building images with Docker 17.05 or higher, you can use [multi-stage builds][] to easily include the `gomplate` binary in your container images.

Use the `COPY` instruction's `--from` flag to accomplish this:

```Dockerfile
...
COPY --from=hairyhenderson/gomplate:stable /gomplate /bin/gomplate
```

Now, `gomplate` will be available in the `/bin` directory inside the container image.

Note that when using `gomplate` with HTTPS-based datasources, you will likely need to install the `ca-certificates` package for your base distribution. Here's an example when using the [`alpine`](https://hub.docker.com/_/alpine) base image:

```Dockerfile
FROM alpine

COPY --from=hairyhenderson/gomplate:stable /gomplate /bin/gomplate
RUN apk add --no-cache ca-certificates
```

## manual install

1. Get the latest `gomplate` for your platform from the [releases][] page
2. Store the downloaded binary somewhere in your path as `gomplate` (or `gomplate.exe`
  on Windows)
3. Make sure it's executable (on Linux/macOS)
4. Test it out with `gomplate --help`!

In other words:

```console
$ curl -o /usr/local/bin/gomplate -sSL https://github.com/hairyhenderson/gomplate/releases/download/<version>/gomplate_<os>-<arch>
$ chmod 755 /usr/local/bin/gomplate
$ gomplate --help
...
```

## install with `go install`

If you're a Go developer, sometimes it's faster to just use `go install` to install `gomplate`:

```console
$ go install github.com/hairyhenderson/gomplate/v4/cmd/gomplate@latest
$ gomplate --help
...
```

(note that this method produces a binary that isn't versioned and may not necessarily work correctly)

## install with `npm`

For some users, especially Node.js developers, using `npm` may be a natural fit.
Even though `gomplate` is written in Go and not Node.js, it can still be installed
with `npm`:

```console
$ npm install -g gomplate
...
```

## install with `tea.xyz`

For some users (including for DEVOPS on [GitHub Actions](https://github.com/marketplace/actions/tea-setup)),
[`tea.xyz`](https://tea.xyz/) maybe be very comfortable, therefore, to install, just:

```console
$ sh <(curl https://tea.xyz) +gomplate.ca sh
$ gomplate --version
...
```

[releases]: https://github.com/hairyhenderson/gomplate/releases
[multi-stage builds]: https://docs.docker.com/develop/develop-images/multistage-build/
[hairyhenderson/gomplate]: https://hub.docker.com/r/hairyhenderson/gomplate/tags/
