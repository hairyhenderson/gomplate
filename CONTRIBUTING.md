# Contributing to gomplate

Thanks for considering a contribution! All contributions are welcome, including
bug reports, bug fixes, new features or even just questions.

## Features

For PRs, please:
- Consider filing an [issue](https://github.com/hairyhenderson/gomplate/issues/new) first, especially if you're not sure if your idea will be accepted.
- Commit messages should follow the repo's [conventional commit](https://www.conventionalcommits.org/en/v1.0.0/) style. See [`release-please-config.json`](./release-please-config.json) for valid commit types.
- [Link to any relevant issues](https://help.github.com/articles/autolinked-references-and-urls/) in the PR
- Add new tests to cover the new code, and make sure all tests pass (`make lint test integration`)
- Please try to conform to the existing code style, and see [Style Guide](#style-guide) for more details

## Bugs

Submit issues to [issue tracker](https://github.com/hairyhenderson/gomplate/issues/).

Any bug fix PRs must also include unit tests and/or integration tests to prevent regression.

If you think you've found a sensitive security issue, please e-mail me before opening an issue: dhenderson@gmail.com. My PGP key is available on Keybase: https://keybase.io/dhenderson/.

## Versioning, API and Deprecation

I try to follow [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html) as much as possible, and as such the version numbers attached to gomplate releases have _semantic meaning_. That is, patch-level releases (e.g. when going from 2.3.0 to 2.3.1) will only contain bug fixes and will not otherwise add or remove features, while minor-level releases (e.g. 2.3.0 to 2.4.0) will only contain new features and will not remove or make breaking changes to features, and major-level releases (e.g. 2.3.0 to 3.0.0) will contain breaking changes (either feature removals, renames, or significant behaviour changes).

When making a change that'll require either a _minor_ or _major_ version bump upon release, the PR should be labeled with the `api/minor` or `api/major` labels.

Note also that _only_ the "Public API" is subject to the semver guarantees. Here's what I consider as "Public API":
1. exported code
  - this includes the behaviour of that code as documented
  - excluding function definitions (in the `funcs/` sub-directory)
  - excluding unit and integration tests
2. template functions
  - the namespaced function names and aliases, as used inside templates (e.g. `math.Add`)
  - the Go functions that implement these functions _may_ change:
    - the function signature may change in compatible ways (i.e. a parameter type widening from `string` to `interface{}`)
3. the `gomplate` command and its command-line arguments

Any semver guarantees only apply between released versions of gomplate. I reserve the right to make breaking changes to new features before these features are delivered in a release.

Sometimes, a deprecation is necessary. I will mark these deprecations in the documentation as soon as possible, and the deprecations may turn into removals (or renames, behaviour changes, etc...) as early as the next major version. However, features may stay deprecated for many major versions without being removed.

## Style Guide

Code style is enforced by [`golangci-lint`](https://github.com/golangci/golangci-lint) during CI builds.

### Template Function Style

Gomplate's code base has grown organically, and so it's full of quirks and inconsistencies. However, there are a few style points that have emerged over the years:

1. Generally, all template functions should take `interface{}` as an input type, and do type conversions where necessary. As an example, see https://github.com/hairyhenderson/gomplate/blob/af961ebe30041acb4e7f94ddd5f7c92372f97111/funcs/math.go#L96..L111
    - In cases where a function takes a slice as input, use `interface{}` as the input, and convert it to a `[]interface{}` if needed
2. Where defaults can be inferred, I prefer allowing a relaxed polymorphic style - see https://github.com/hairyhenderson/gomplate/blob/af961ebe30041acb4e7f94ddd5f7c92372f97111/funcs/crypto.go#L118..L133 as an example
    - tl;dr: I take `...interface{}` as a single input type, and do type conversion and argument inference based on the number of arguments provided
    - this only works when certain args can be considered "optional"
3. Pipelining (i.e. `{{ "result of some action" | some.Function }}`) is powerful and it's good to make sure that functions with multiple arguments specify them in a sane order for pipelining.
    - When you're unsure about argument order, prefer the order that will make pipelining easier:
      ```
      {{ "one,two,three" | strings.Split "," }} <- this is far more natural
      ```
    - Don't depend too much on existing functions as a guide - there's already a lot of inconsistency there!
