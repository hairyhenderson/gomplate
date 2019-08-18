---
title: Usage
weight: 11
menu: main
---

The simplest usage of `gomplate` is to just replace environment
variables. All environment variables are available by referencing [`.Env`](../syntax/#env)
(or [`getenv`](../functions/env/#getenv)) in the template.

The template is read from standard in, and written to standard out.

Use it like this:

```console
$ echo "Hello, {{.Env.USER}}" | gomplate
Hello, hairyhenderson
```

## Commandline Arguments

### `--file`/`-f`, `--in`/`-i`, and `--out`/`-o`

By default, `gomplate` will read from `Stdin` and write to `Stdout`. This behaviour can be changed.

- Use `--file`/`-f` to use a specific input template file. The special value `-` means `Stdin`.
- Use `--out`/`-o` to save output to file. The special value `-` means `Stdout`.
- Use `--in`/`-i` if you want to set the input template right on the commandline. This overrides `--file`. Because of shell command line lengths, it's probably not a good idea to use a very long value with this argument.

#### Multiple inputs

You can specify multiple `--file` and `--out` arguments. The same number of each much be given. This allows `gomplate` to process multiple templates _slightly_ faster than invoking `gomplate` multiple times in a row.

### `--input-dir` and `--output-dir`

For processing multiple templates in a directory you can use `--input-dir` and `--output-dir` together. In this case all files in input directory will be processed as templates and the resulting files stored in `--output-dir`. The output directory will be created if it does not exist and the directory structure of the input directory will be preserved.

You can use the [`--exclude`](#exclude) argument and/or a [`.gomplateignore`](#ignorefile) file to exclude some of the files in the input directory.

Example:

```bash
# Process all files in directory "templates" with the datasource given
# and store the files with the same directory structure in "config"
gomplate --input-dir=templates --output-dir=config --datasource config=config.yaml
```

### `--output-map`

Sometimes a 1-to-1 mapping betwen input filenames and output filenames is not desirable. For these cases, you can supply a template string as the argument to `--output-map`. The template string is interpreted as a regular gomplate template, and all datasources and external nested templates are available to the output map template.

A new [context][] is provided, with the input filename is available at `.in`, and the original context is available at `.ctx`. For convenience, any context keys not conflicting with `in` or `ctx` are also copied.

All whitespace on the left or right sides of the output is trimmed.

For example, given an input directory `in/` containing files with the extension `.yaml.tmpl`, if we want to rename those to `.yaml`:

```console
$ gomplate --input-dir=in/ --output-map='out/{{ .in | strings.ReplaceAll ".yaml.tmpl" ".yaml" }}'
```

#### Referencing complex output map template files

It may be useful to store more complex output map templates in a file. This can be done with [external templates][].

Consider a template `out.t`:

```
{{- /* .in may contain a directory name - we want to preserve that */ -}}
{{ $f := filepath.Base .in -}}
out/{{ .in | strings.ReplaceAll $f (index .filemap $f) }}.out
```

And a datasource `filemap.json`:

```json
{ "eins.txt": "uno", "deux.txt": "dos" }
```

We can blend these two together:

```console
$ gomplate -t out=out.t -c filemap.json --input-dir=in --output-map='{{ template "out" }}'
```

### `--chmod`

By default, output files are created with the same file mode (permissions) as input files. If desired, the `--chmod` option can be used to override this behaviour, and set the output file mode explicitly. This can be useful for creating executable scripts or ensuring write permissions.

The value must be an octal integer in the standard UNIX `chmod` format, i.e. `644` to indicate that owner gets read+write, group gets read-only, and others get read-only permissions. See the [`chmod(1)` man page](https://linux.die.net/man/1/chmod) for more details.

**Note:** `--chmod` is not currently supported on Windows. Behaviour is undefined, but will likely not change file permissions at all.

### `--exclude` and `--include`

When using the [`--input-dir`](#input-dir-and-output-dir) argument, it can be useful to filter which files are processed. You can use `--exclude` and `--include` to achieve this. The `--exclude` flag takes a [`.gitignore`][]-style pattern, and any files matching the pattern will be excluded. The `--include` flag is effectively the opposite of `--exclude`. You can also repeat the arguments to provide a series of patterns to be excluded/included.

Patterns provided with `--exclude`/`--include` are matched relative to the input directory.

_Note:_ These patterns are _not_ treated as filesystem globs, and so a pattern like `/foo/bar.json` will match relative to the input directory, not the root of the filesystem as they may appear!

Examples:

```console
$ gomplate --exclude example/** --exclude *.png --input-dir in/ --output-dir out/
```

This will stop all files in the `in/example` directory from being processed, as well as all `.png` files in the `in/` directory.

```console
$ gomplate --include *.tmpl --exclude foo*.tmpl --input-dir in/ --output-dir out/
```

This will cause only files ending in `.tmpl` to be processed, except for files with names beginning with `foo`: `template.tmpl` will be included, but `foo-template.tmpl` will not.

#### `.gomplateignore` files

You can also use a file named `.gomplateignore` containing one exclude pattern on each line. This has the same syntax as a [`.gitignore`][] file.
When processing sub-directories, `.gomplateignore` files in the parent directory are also considered. Patterns are matched relative to the location of the `.gomplateignore` file.

### `--datasource`/`-d`

Add a data source in `name=URL` form. Specify multiple times to add multiple sources. The data can then be used by the [`datasource`](../functions/data/#datasource) and [`include`](../functions/data/#include) functions.

Data sources referenced in this way are lazy-loaded: they will not be read until the template is parsed and a `datasource` or `include` function is encountered.

See [Datasources](../datasources) for full details.

A few different forms are valid:
- `mydata=file:///tmp/my/file.json`
  - Create a data source named `mydata` which is read from `/tmp/my/file.json`. This form is valid for any file in any path.
- `mydata=file.json`
  - Create a data source named `mydata` which is read from `file.json` (in the current working directory). This form is only valid for files in the current directory.
- `mydata.json`
  - This form infers the name from the file name (without extension). Only valid for files in the current directory.

### `--context`/`c`

Add a data source in `name=URL` form, and make it available in the [default context][] as `.<name>`. The special name `.` (period) can be used to override the entire default context.

Data sources referenced with `--context` will be immediately loaded before gomplate processes the template. This is in contrast to the `--datasource` behaviour, which lazy-loads data while processing the template.

All other rules for the [`--datasource`/`-d`](#datasource-d) flag apply.

Examples:

```console
$ gomplate --context post=https://jsonplaceholder.typicode.com/posts/2 -i 'post title is: {{ .post.title }}'
post title is: qui est esse
```

```console
$ gomplate -c .=http://xkcd.com/info.0.json -i '<a href="{{ .img }}">{{ .title }}</a>'
<a href="https://imgs.xkcd.com/comics/diploma_legal_notes.png">Diploma Legal Notes</a>
```

### Overriding the template delimiters

Sometimes it's necessary to override the default template delimiters (`{{`/`}}`).
Use `--left-delim`/`--right-delim` or set `$GOMPLATE_LEFT_DELIM`/`$GOMPLATE_RIGHT_DELIM`.

### `--template`/`-t`

Add a nested template that can be referenced by the main input template(s) with the [`template`](https://golang.org/pkg/text/template/#hdr-Actions) built-in or the functions in the [`tmpl`](../functions/tmpl/) namespace. Specify multiple times to add multiple template references.

A few different forms are valid:

- `--template mytemplate.t`
  - References a file `mytemplate.t` in the current working directory.
  - It will be available as a template named `mytemplate.t`:
    ```console
    $ gomplate --template helloworld.tmpl -i 'here are the contents of the template: [ {{ template "helloworld.tmpl" }} ]'
    here are the contents of the template: [ hello, world! ]
    ```
- `--template path/to/mytemplate.t`
  - References a file `mytemplate.t` in the path `path/to/`.
  - It will be available as a template named `path/to/mytemplate.t`:
    ```console
    $ gomplate --template foo/bar/helloworld.tmpl -i 'here are the contents of the template: [ {{ template "foo/bar/helloworld.tmpl" }} ]'
    here are the contents of the template: [ hello, world! ]
    ```
- `--template path/to/`
  - Makes available all files in the path `path/to/`.
  - Any files within this path can be referenced:
    ```console
    $ gomplate --template foo/bar/ -i 'here are the contents of the template: [ {{ template "foo/bar/helloworld.tmpl" }} ]'
    here are the contents of the template: [ hello, world! ]
    ```
- `--template alias=path/to/mytemplate.t`
  - References a file `mytemplate.t` in the path `path/to/`
  - It will be available as a template named `alias`:
    ```console
    $ gomplate --template t=foo/bar/helloworld.tmpl -i 'here are the contents of the template: [ {{ template "t" }} ]'
    here are the contents of the template: [ hello, world! ]
    ```
- `--template alias=path/to/`
  - Makes available all files in the path `path/to/`.
  - Any files within this path can be referenced, with the path replaced with `alias`:
    ```console
    $ gomplate --template dir=foo/bar/ -i 'here are the contents of the template: [ {{ template "dir/helloworld.tmpl" }} ]'
    here are the contents of the template: [ hello, world! ]
    ```

## Post-template command execution

Gomplate can launch other commands when template execution is successful. Simply
add the command to the command-line after a `--` argument:

```console
$ gomplate -i 'hello world' -o out.txt -- cat out.txt
hello world
```

## Suppressing empty output

Sometimes it can be desirable to suppress empty output (i.e. output consisting of only whitespace). To do so, set `GOMPLATE_SUPPRESS_EMPTY=true` in your environment:

```console
$ export GOMPLATE_SUPPRESS_EMPTY=true
$ gomplate -i '{{ print "   \n" }}' -o out
$ cat out
cat: out: No such file or directory
```

[default context]: ../syntax/#the-context
[context]: ../syntax/#the-context
[external templates]: ../syntax/#external-templates
[`.gitignore`]: https://git-scm.com/docs/gitignore
