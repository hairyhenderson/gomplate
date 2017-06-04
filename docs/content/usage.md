---
title: Usage
weight: 0
---

The simplest usage of `gomplate` is to just replace environment
variables. All environment variables are available by referencing [`.Env`](../syntax/#about-env)
(or [`getenv`](../functions/#getenv)) in the template.

The template is read from standard in, and written to standard out.

Use it like this:

```console
$ echo "Hello, {{.Env.USER}}" | gomplate
Hello, hairyhenderson
```

# Commandline Arguments

## `--file`/`-f`, `--in`/`-i`, and `--out`/`-o`

By default, `gomplate` will read from `Stdin` and write to `Stdout`. This behaviour can be changed.

- Use `--file`/`-f` to use a specific input template file. The special value `-` means `Stdin`.
- Use `--out`/`-o` to save output to file. The special value `-` means `Stdout`.
- Use `--in`/`-i` if you want to set the input template right on the commandline. This overrides `--file`. Because of shell command line lengths, it's probably not a good idea to use a very long value with this argument.

### Multiple inputs

You can specify multiple `--file` and `--out` arguments. The same number of each much be given. This allows `gomplate` to process multiple templates _slightly_ faster than invoking `gomplate` multiple times in a row.

### `--input-dir` and `--output-dir`

For processing multiple templates in a directory you can use `--input-dir` and `--output-dir` together. In this case all files in input directory will be processed as templates and the resulting files stored in `--output-dir`. The output directory will be created if it does not exist and the directory structure of the input directory will be preserved.  

Example:

```bash
 # Process all files in directory "templates" with the datasource given
 # and store the files with the same directory structure in "config"
gomplate --input-dir=templates --output-dir=config --datasource config=config.yaml
```

## `--datasource`/`-d`

Add a data source in `name=URL` form. Specify multiple times to add multiple sources. The data can then be used by the [`datasource`](../functions/#datasource) and [`include`](../functions/#include) functions.

A few different forms are valid:
- `mydata=file:///tmp/my/file.json`
  - Create a data source named `mydata` which is read from `/tmp/my/file.json`. This form is valid for any file in any path.
- `mydata=file.json`
  - Create a data source named `mydata` which is read from `file.json` (in the current working directory). This form is only valid for files in the current directory.
- `mydata.json`
  - This form infers the name from the file name (without extension). Only valid for files in the current directory.

## Overriding the template delimiters

Sometimes it's necessary to override the default template delimiters (`{{`/`}}`).
Use `--left-delim`/`--right-delim` or set `$GOMPLATE_LEFT_DELIM`/`$GOMPLATE_RIGHT_DELIM`.
