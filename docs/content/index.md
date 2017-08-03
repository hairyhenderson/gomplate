---
title: gomplate
type: index
weight: 1
menu:
  main:
    name: About
---

A [Go template](https://golang.org/pkg/text/template/)-based CLI tool. `gomplate` can be used as an alternative to
[`envsubst`](https://www.gnu.org/software/gettext/manual/html_node/envsubst-Invocation.html) but also supports
additional template datasources such as: JSON, YAML, AWS EC2 metadata, [BoltDB](https://github.com/boltdb/bolt),
[Hashicorp Consul](https://www.consul.io/) and [Hashicorp Vault](https://www.vaultproject.io/) secrets.

I really like `envsubst` for use as a super-minimalist template processor. But its simplicity is also its biggest flaw: it's all-or-nothing with shell-like variables.

Gomplate is an alternative that will let you process templates which also include shell-like variables. Also there are some useful built-in functions that can be used to make templates even more expressive.


_Please report any bugs found in the [issue tracker](https://github.com/hairyhenderson/gomplate/issues/)._

{{< note title="Note" >}}
This documentation is still in the process of being migrated out of the
[README](https://github.com/hairyhenderson/gomplate/tree/master/README.md), so
expect some inconsistencies! If you want to help, [PRs and issues are welcome!](https://github.com/hairyhenderson/gomplate/issues/new)
{{< /note >}}

## License

[The MIT License](http://opensource.org/licenses/MIT)

Copyright (c) 2016-2017 Dave Henderson
