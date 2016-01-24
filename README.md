[![Build Status][circleci-image]][circleci-url]

# gomplate

A simple [Go template](https://golang.org/pkg/text/template/)-based alternative to [`envsubst`](https://www.gnu.org/software/gettext/manual/html_node/envsubst-Invocation.html).

I really like `envsubst` for use as a super-minimalist template processor. But its simplicity is also its biggest flaw: it's all-or-nothing with shell-like variables.

Gomplate is an alternative that will let you process templates which also include shell-like variables.

## Usage

At the moment, `gomplate` just replaces environment variables. All environment variables are available by referencing `.Env` in the template.

The template is read from standard in, and written to standard out.

Use it like this:

```console
$ echo "Hello, {{.Env.USER}}" | gomplate
Hello, hairyhenderson
```

## License

[The MIT License](http://opensource.org/licenses/MIT)

Copyright (c) 2016 Dave Henderson

[circleci-image]: https://img.shields.io/circleci/project/hairyhenderson/gomplate.svg?style=flat
[circleci-url]: https://circleci.com/gh/hairyhenderson/gomplate
