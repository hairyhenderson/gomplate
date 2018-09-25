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

Example:

```bash
 # Process all files in directory "templates" with the datasource given
 # and store the files with the same directory structure in "config"
gomplate --input-dir=templates --output-dir=config --datasource config=config.yaml
```

### `--chmod`

By default, output files are created with the same file mode (permissions) as input files. If desired, the `--chmod` option can be used to override this behaviour, and set the output file mode explicitly. This can be useful for creating executable scripts or ensuring write permissions.

The value must be an octal integer in the standard UNIX `chmod` format, i.e. `644` to indicate that owner gets read+write, group gets read-only, and others get read-only permissions. See the [`chmod(1)` man page](https://linux.die.net/man/1/chmod) for more details.

### `--exclude`

To prevent certain files from being processed, you can use `--exclude`. It takes a glob, and any files matching that glob will not be included.

Example:

```console
$ gomplate --exclude example/** --exclude *.png
```

This will stop all files in the example folder from being processed, as well as all `.png` files in the current folder.

You can also chain the flag to build up a series of globs to be excluded.

### `--datasource`/`-d`

Add a data source in `name=URL` form. Specify multiple times to add multiple sources. The data can then be used by the [`datasource`](../functions/data/#datasource) and [`include`](../functions/data/#include) functions.

See [Datasources](../datasources) for full details.

A few different forms are valid:
- `mydata=file:///tmp/my/file.json`
  - Create a data source named `mydata` which is read from `/tmp/my/file.json`. This form is valid for any file in any path.
- `mydata=file.json`
  - Create a data source named `mydata` which is read from `file.json` (in the current working directory). This form is only valid for files in the current directory.
- `mydata.json`
  - This form infers the name from the file name (without extension). Only valid for files in the current directory.

### Overriding the template delimiters

Sometimes it's necessary to override the default template delimiters (`{{`/`}}`).
Use `--left-delim`/`--right-delim` or set `$GOMPLATE_LEFT_DELIM`/`$GOMPLATE_RIGHT_DELIM`.

### `--template`/`-t`

Add a nested template that can be referenced by the main input template(s) with the [`template`](https://golang.org/pkg/text/template/#hdr-Actions) built-in. Specify multiple times to add multiple template references.

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
