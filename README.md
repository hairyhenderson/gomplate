<img src="docs/static/images/gomplate.png" width="512px" alt="gomplate logo"/>

_Read the docs at [gomplate.hairyhenderson.ca][docs-url]._

[![Build Status][circleci-image]][circleci-url]
[![Windows Build][appveyor-image]][appveyor-url]
[![Go Report Card][reportcard-image]][reportcard-url]
[![Codebeat Status][codebeat-image]][codebeat-url]
[![Coverage][gocover-image]][gocover-url]
[![Total Downloads][gh-downloads-image]][gh-downloads-url]
[![CII Best Practices][cii-bp-image]][cii-bp-url]
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fhairyhenderson%2Fgomplate.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fhairyhenderson%2Fgomplate?ref=badge_shield)

[![hairyhenderson/gomplate on DockerHub][dockerhub-image]][dockerhub-url]
[![DockerHub Stars][dockerhub-stars-image]][dockerhub-url]
[![DockerHub Pulls][dockerhub-pulls-image]][dockerhub-url]
[![DockerHub Image Layers][microbadger-layers-image]][microbadger-url]
[![DockerHub Latest Version ][microbadger-version-image]][microbadger-url]
[![DockerHub Latest Commit][microbadger-commit-image]][microbadger-url]

[![Install Docs][install-docs-image]][install-docs-url]

A [Go template](https://golang.org/pkg/text/template/)-based CLI tool. `gomplate` can be used as an alternative to
[`envsubst`](https://www.gnu.org/software/gettext/manual/html_node/envsubst-Invocation.html) but also supports
additional template datasources such as: JSON, YAML, AWS EC2 metadata, [BoltDB](https://github.com/boltdb/bolt),
[Hashicorp Consul](https://www.consul.io/) and [Hashicorp Vault](https://www.vaultproject.io/) secrets.

I really like `envsubst` for use as a super-minimalist template processor. But its simplicity is also its biggest flaw: it's all-or-nothing with shell-like variables.

Gomplate is an alternative that will let you process templates which also include shell-like variables. Also there are some useful built-in functions that can be used to make templates even more expressive.

Read more documentation at [gomplate.hairyhenderson.ca][docs-url]!

_Please report any bugs found in the [issue tracker](https://github.com/hairyhenderson/gomplate/issues/)._

## Releasing

Right now the release process is semi-automatic.

1. Create a release tag: `git tag -a v0.0.9 -m "Releasing v0.9.9" && git push --tags`
2. Build binaries & compress most of them: `make build-release`
3. Create a release in [github](https://github.com/hairyhenderson/gomplate/releases)!

## License

[The MIT License](http://opensource.org/licenses/MIT)

Copyright (c) 2016-2018 Dave Henderson

[circleci-image]: https://circleci.com/gh/hairyhenderson/gomplate/tree/master.svg?style=shield
[circleci-url]: https://circleci.com/gh/hairyhenderson/gomplate/tree/master
[appveyor-image]: https://ci.appveyor.com/api/projects/status/eymky02f5snclyxp/branch/master?svg=true
[appveyor-url]: https://ci.appveyor.com/project/hairyhenderson/gomplate/branch/master
[reportcard-image]: https://goreportcard.com/badge/github.com/hairyhenderson/gomplate
[reportcard-url]: https://goreportcard.com/report/github.com/hairyhenderson/gomplate
[codebeat-image]: https://codebeat.co/badges/39ed2148-4b86-4d1e-8526-25f60e159ba1
[codebeat-url]: https://codebeat.co/projects/github-com-hairyhenderson-gomplate
[gocover-image]: https://gocover.io/_badge/github.com/hairyhenderson/gomplate
[gocover-url]: https://gocover.io/github.com/hairyhenderson/gomplate
[gh-downloads-image]: https://img.shields.io/github/downloads/hairyhenderson/gomplate/total.svg
[gh-downloads-url]: https://github.com/hairyhenderson/gomplate/releases

[cii-bp-image]: https://bestpractices.coreinfrastructure.org/projects/337/badge
[cii-bp-url]: https://bestpractices.coreinfrastructure.org/projects/337

[dockerhub-image]: https://img.shields.io/badge/docker-ready-blue.svg
[dockerhub-url]: https://hub.docker.com/r/hairyhenderson/gomplate
[dockerhub-stars-image]: https://img.shields.io/docker/stars/hairyhenderson/gomplate.svg
[dockerhub-pulls-image]: https://img.shields.io/docker/pulls/hairyhenderson/gomplate.svg

[microbadger-version-image]: https://images.microbadger.com/badges/version/hairyhenderson/gomplate.svg
[microbadger-layers-image]: https://images.microbadger.com/badges/image/hairyhenderson/gomplate.svg
[microbadger-commit-image]: https://images.microbadger.com/badges/commit/hairyhenderson/gomplate.svg
[microbadger-url]: https://microbadger.com/image/hairyhenderson/gomplate

[docs-url]: https://gomplate.hairyhenderson.ca
[install-docs-image]: https://img.shields.io/badge/install-docs-blue.svg
[install-docs-url]: https://gomplate.hairyhenderson.ca/installing

[![Analytics](https://ga-beacon.appspot.com/UA-82637990-1/gomplate/README.md?pixel)](https://github.com/igrigorik/ga-beacon)


[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fhairyhenderson%2Fgomplate.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fhairyhenderson%2Fgomplate?ref=badge_large)