---
title: Configuration
weight: 12
menu: main
---

In addition to [command-line arguments][], gomplate supports the use of
configuration files to control its behaviour.

Using a file for configuration can be useful especially when rendering templates
that use multiple datasources, plugins, nested templates, etc... In situations
where teams share templates, it can be helpful to commit config files into the
team's source control system.

By default, gomplate will look for a file `.gomplate.yaml` in the current working
diretory, but this path can be altered with the [`--config`](../usage/#config)
command-line argument, or the `GOMPLATE_CONFIG` environment variable.

### Configuration precedence

[Command-line arguments][] will always take precedence over settings in a config
file. In the cases where configuration can be altered with an environment
variable, the config file will take precedence over environment variables.

So, if the `leftDelim` setting is configured in 3 ways:

```console
$ export GOMPLATE_LEFT_DELIM=::
$ echo "leftDelim: ((" > .gomplate.yaml
$ gomplate --left-delim "<<"
```

The delimiter will be `<<`.

## File format

Currently, gomplate supports config files written in [YAML][] syntax, though other
structured formats may be supported in future (please [file an issue][] if this
is important to you!)

Roughly all of the [command-line arguments][] are able to be set in a config
file, with the exception of `--help`, `--verbose`, and `--version`. Some 
environment variable based settings not configurable on the command-line are
also supported in config files.

Most of the configuration names are similar, though instead of using `kebab-case`,
multi-word names are rendered as `camelCase`.

Here is an example of a simple config file:

```yaml
inputDir: in/
outputDir: out/

datasources:
  local:
    url: file:///tmp/data.json
  remote:
    url: https://example.com/api/v1/data
    header:
      Authorization: ["Basic aGF4MHI6c3dvcmRmaXNoCg=="]

plugins:
  dostuff: /usr/local/bin/stuff.sh
```

## `chmod`

See [`--chmod`](../usage/#chmod).

Sets the output file mode.

## `context`

See [`--context`](../usage/#context-c).

Add data sources to the default context. This is a nested structure that
includes the URL for the data source and the optional HTTP header to send.

For example:

```yaml
context:
  data:
    url: https://example.com/api/v1/data
    header:
      Authorization: ["Basic aGF4MHI6c3dvcmRmaXNoCg=="]
  stuff:
    url: stuff.yaml
```

This adds two datasources to the context: `data` and `stuff`, and when the `data`
source is retrieved, an `Authorization` header will be sent with the given value.

Note that the `.` name can also be used to set the entire context:

```yaml
context:
  .:
    url: data.toml
```

## `datasources`

See [`--datasource`](../usage/#datasource-d).

Define data sources. This is a nested structure that includes the URL for the data
source and the optional HTTP header to send.

For example:

```yaml
datasources:
  data:
    url: https://example.com/api/v1/data
    header:
      Authorization: ["Basic aGF4MHI6c3dvcmRmaXNoCg=="]
  stuff:
    url: stuff.yaml
```

This defines two datasources: `data` and `stuff`, and when the `data`
source is used, an `Authorization` header will be sent with the given value.

## `excludes`

See [`--exclude` and `--include`](../usage/#exclude-and-include).

This is an array of exclude patterns, used in conjunction with [`inputDir`](#inputdir).
Note that there is no `includes`, instead you can specify negative
exclusions by prefixing the patterns with `!`.

```yaml
excludes:
  - '*.txt'
  - '!include-this.txt'
```

This will skip all files with the extension `.txt`, except for files named
`include-this.txt`, which will be processed.

## `execPipe`

See [`--exec-pipe`](../usage/#exec-pipe).

Use the rendered output as the [`postExec`](#postexec) command's standard input.

Must be used in conjuction with [`postExec`](#postexec), and will override
any [`outputFiles`](#outputfiles) settings.

## `experimental`

See [`--experimental`](../usage/#experimental). Can also be set with the `GOMPLATE_EXPERIMENTAL=true` environment variable.

Some functions and features are provided for early feedback behind the `experimental` configuration option. These features may change before being permanently enabled, and [feedback](https://github.com/hairyhenderson/gomplate/issues/new) is requested from early adopters!

Experimental functions are marked in the documentation with an _"(experimental)"_ annotation.

```yaml
experimental: true
```

## `in`

See [`--in`/`-i`](../usage/#file-f-in-i-and-out-o).

Provide the input template inline. Note that unlike the `--in`/`-i` commandline
argument, there are no shell-imposed length limits.

A simple example:
```yaml
in: hello to {{ .Env.USER }}
```

A multi-line example (see https://yaml-multiline.info/ for more about multi-line
string syntax in YAML):

```yaml
in: |
  A longer multi-line
  document:
  {{- range .foo }}
  {{ .bar }}
  {{ end }}
```

May not be used with `inputDir` or `inputFiles`.

## `inputDir`

See [`--input-dir`](../usage/#input-dir-and-output-dir).

The directory containing input template files. Must be used with 
[`outputDir`](#outputdir) or [`outputMap`](#outputmap). Can also be used with [`excludes`](#excludes).

```yaml
inputDir: templates/
outputDir: out/
```

May not be used with `in` or `inputFiles`.

## `inputFiles`

See [`--file`/`-f`](../usage/#file-f-in-i-and-out-o).

An array of input template paths. The special value `-` means `Stdin`. Multiple
values can be set, but there must be a corresponding number of `outputFiles`
entries present.

```yaml
inputFiles:
  - first.tmpl
  - second.tmpl
outputFiles:
  - first.out
  - second.out
```

Flow style can be more compact:

```yaml
inputFiles: ['-']
outputFiles: ['-']
```

May not be used with `in` or `inputDir`.

## `leftDelim`

See [`--left-delim`](../usage/#overriding-the-template-delimiters).

Overrides the left template delimiter.

```yaml
leftDelim: '%{'
```

## `outputDir`

See [`--output-dir`](../usage/#input-dir-and-output-dir).

The directory to write rendered output files. Must be used with 
[`inputDir`](#inputdir).

If the directory is missing, it will be created with the same permissions as the
`inputDir`.

```yaml
inputDir: templates/
outputDir: out/
```

May not be used with `outputFiles`.

## `outputFiles`

See [`--out`/`-o`](../usage/#file-f-in-i-and-out-o).

An array of output file paths. The special value `-` means `Stdout`. Multiple
values can be set, but there must be a corresponding number of `inputFiles`
entries present.

If any of the parent directories are missing, they will be created with the same
permissions as the input directories.

```yaml
inputFiles:
  - first.tmpl
  - second.tmpl
outputFiles:
  - first.out
  - second.out
```

Can also be used with [`in`](#in):

```yaml
in: >-
  hello,
  world!
outputFiles: [ hello.txt ]
```

May not be used with `inputDir`.

## `outputMap`

See [`--output-map`](../usage/#output-map).

Must be used with [`inputDir`](#inputdir).

```yaml
inputDir: in/
outputMap: |
  out/{{ .in | strings.ReplaceAll ".yaml.tmpl" ".yaml" }}
```

## `plugins`

See [`--plugin`](../usage/#plugin).

A map that configures custom functions for use in the templates. The key is the
name of the function, and the value configures the plugin. The value is a map
containing the command (`cmd`) and the options `pipe` (boolean) and `timeout`
(duration).

Alternatively, the value can be a string, which sets `cmd`.

```yaml
in: '{{ "hello world" | figlet | lolcat }}'
plugins:
  figlet:
    cmd: /usr/local/bin/figlet
    pipe: true
    timeout: 1s
  lolcat: /home/hairyhenderson/go/bin/lolcat
```

### `cmd`

The path to the plugin executable (or script) to run.

### `pipe`

Whether to pipe the final argument of the template function to the plugin's
Stdin, or provide as a separate argument.

For example, given a `myfunc` plugin with a `cmd` of `/bin/myfunc`:

With this template:
```
{{ print "bar" | myfunc "foo" }}
```

If `pipe` is `true`, the plugin executable will receive the input `"bar"` as its
Stdin, like this shell command:

```console
$ echo -n "bar" | /bin/myfunc "foo"
```

If `pipe` is `false` (the default), the plugin executable will receive the
input `"bar"` as its last argument, like this shell command:

```console
$ /bin/myfunc "foo" "bar"
```

_Note:_ in a chained pipeline (e.g. `{{ foo | bar }}`), the result of each
command is passed as the final argument of the next, and so the template above
could be written as `{{ myfunc "foo" "bar" }}`.

### `timeout`

The plugin's timeout. After this time, the command will be terminated and the
template function will return an error. The value must be a valid
[duration][] such as `1s`, `1m`, `1h`,

The default is `5s`.

## `pluginTimeout`

See [`--plugin`](../usage/#plugin).

Sets the timeout for all configured plugins. Overrides the default of `5s`.
After this time, plugin commands will be killed. The value must be a valid
[duration][] such as `10s` or `3m`.

```yaml
plugins:
  figlet: /usr/local/bin/figlet
pluginTimeout: 500ms
```

## `postExec`

See [post-template command execution](../usage/#post-template-command-execution).

Configures a command to run after the template is rendered.

See also [`execPipe`](#execpipe) for piping output directly into the `postExec` command.

## `rightDelim`

See [`--right-delim`](../usage/#overriding-the-template-delimiters).

Overrides the right template delimiter.

```yaml
rightDelim: '))'
```

## `suppressEmpty`

See _[Suppressing empty output](../usage/#suppressing-empty-output)_

Suppresses empty output (i.e. output consisting of only whitespace). Can also be set with the `GOMPLATE_SUPPRESS_EMPTY` environment variable.

```yaml
suppressEmpty: true
```

## `templates`

See [`--template`/`-t`](../usage/#template-t).

An array of template references. Can be just a path or an alias and a path:

```yaml
templates:
  - t=foo/bar/helloworld.tmpl
  - templatedir/
  - dir=foo/bar/
  - mytemplate.t
```

[command-line arguments]: ../usage
[file an issue]: https://github.com/hairyhenderson/gomplate/issues/new
[YAML]: http://yaml.org
[duration]: (../functions/time/#time-parseduration)
