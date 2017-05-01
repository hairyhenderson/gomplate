# Change Log

## [v1.6.0](https://github.com/hairyhenderson/gomplate/tree/v1.6.0) (2017-05-01)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.5.1...v1.6.0)

**Implemented enhancements:**

- Support for bulk operations [\#117](https://github.com/hairyhenderson/gomplate/issues/117)
- Authentication for HTTP/HTTPS datasources [\#113](https://github.com/hairyhenderson/gomplate/issues/113)
- Make all secrets settable via files [\#106](https://github.com/hairyhenderson/gomplate/issues/106)
- Adding ds alias for datasource function [\#129](https://github.com/hairyhenderson/gomplate/pull/129) ([hairyhenderson](https://github.com/hairyhenderson))
- Add --input-dir and --output-dir as options [\#119](https://github.com/hairyhenderson/gomplate/pull/119) ([rhuss](https://github.com/rhuss))
- Adding more ways to specify input/output [\#114](https://github.com/hairyhenderson/gomplate/pull/114) ([hairyhenderson](https://github.com/hairyhenderson))

**Fixed bugs:**

- Open datasource file in read-only mode [\#126](https://github.com/hairyhenderson/gomplate/pull/126) ([rhuss](https://github.com/rhuss))

**Merged pull requests:**

- Migrating to spf13/cobra for commandline processing [\#128](https://github.com/hairyhenderson/gomplate/pull/128) ([hairyhenderson](https://github.com/hairyhenderson))
- Updating vendored deps [\#127](https://github.com/hairyhenderson/gomplate/pull/127) ([hairyhenderson](https://github.com/hairyhenderson))
- Removing integration test dependency on internet access [\#121](https://github.com/hairyhenderson/gomplate/pull/121) ([hairyhenderson](https://github.com/hairyhenderson))
- Updating vendored deps \(aws-sdk-go and go-yaml\) [\#120](https://github.com/hairyhenderson/gomplate/pull/120) ([hairyhenderson](https://github.com/hairyhenderson))
- Fix readme ToC link to `--datasource-d` [\#118](https://github.com/hairyhenderson/gomplate/pull/118) ([jamiemjennings](https://github.com/jamiemjennings))
- Support arbitrary headers with HTTP datasources [\#115](https://github.com/hairyhenderson/gomplate/pull/115) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding some very basic integration tests [\#112](https://github.com/hairyhenderson/gomplate/pull/112) ([hairyhenderson](https://github.com/hairyhenderson))

## [v1.5.1](https://github.com/hairyhenderson/gomplate/tree/v1.5.1) (2017-03-23)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.5.0...v1.5.1)

**Implemented enhancements:**

- Support Vault authentication on HTTPS datasource [\#54](https://github.com/hairyhenderson/gomplate/issues/54)
- Supporting \*\_FILE env vars for vault datasource credentials [\#107](https://github.com/hairyhenderson/gomplate/pull/107) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding support for AppRole vault auth backend [\#105](https://github.com/hairyhenderson/gomplate/pull/105) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding support for userpass vault auth backend [\#104](https://github.com/hairyhenderson/gomplate/pull/104) ([hairyhenderson](https://github.com/hairyhenderson))
- Allow custom auth backend mount point for app-id backend [\#103](https://github.com/hairyhenderson/gomplate/pull/103) ([hairyhenderson](https://github.com/hairyhenderson))

**Fixed bugs:**

- Awful performance rendering templates with `ec2tag` function in non-aws environments [\#110](https://github.com/hairyhenderson/gomplate/issues/110)
- Performance fixes for running ec2tag in non-aws environments [\#111](https://github.com/hairyhenderson/gomplate/pull/111) ([hairyhenderson](https://github.com/hairyhenderson))

**Merged pull requests:**

- Clean up vault auth strategies code [\#130](https://github.com/hairyhenderson/gomplate/pull/130) ([hairyhenderson](https://github.com/hairyhenderson))

## [v1.5.0](https://github.com/hairyhenderson/gomplate/tree/v1.5.0) (2017-03-07)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.4.0...v1.5.0)

**Implemented enhancements:**

- Allow setting custom delimiters [\#100](https://github.com/hairyhenderson/gomplate/issues/100)
- Allow overriding the template delimiters [\#102](https://github.com/hairyhenderson/gomplate/pull/102) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding 'has' func to determine if an object has a named key [\#101](https://github.com/hairyhenderson/gomplate/pull/101) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding toJSON and toYAML functions [\#99](https://github.com/hairyhenderson/gomplate/pull/99) ([hairyhenderson](https://github.com/hairyhenderson))

## [v1.4.0](https://github.com/hairyhenderson/gomplate/tree/v1.4.0) (2017-03-03)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.3.0...v1.4.0)

**Implemented enhancements:**

- Adding more functions from the strings package [\#96](https://github.com/hairyhenderson/gomplate/pull/96) ([hairyhenderson](https://github.com/hairyhenderson))

**Merged pull requests:**

- shutting up golint [\#97](https://github.com/hairyhenderson/gomplate/pull/97) ([hairyhenderson](https://github.com/hairyhenderson))
- Putting vendor/ in repo [\#95](https://github.com/hairyhenderson/gomplate/pull/95) ([hairyhenderson](https://github.com/hairyhenderson))

## [v1.3.0](https://github.com/hairyhenderson/gomplate/tree/v1.3.0) (2017-02-03)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.2.4...v1.3.0)

**Implemented enhancements:**

- Adding datasourceExists function [\#94](https://github.com/hairyhenderson/gomplate/pull/94) ([hairyhenderson](https://github.com/hairyhenderson))

**Closed issues:**

- Crash when datasource is not specified [\#93](https://github.com/hairyhenderson/gomplate/issues/93)

## [v1.2.4](https://github.com/hairyhenderson/gomplate/tree/v1.2.4) (2017-01-13)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.2.3...v1.2.4)

**Merged pull requests:**

- Building a slim macOS binary too [\#92](https://github.com/hairyhenderson/gomplate/pull/92) ([hairyhenderson](https://github.com/hairyhenderson))
- Vendoring dependencies with glide [\#91](https://github.com/hairyhenderson/gomplate/pull/91) ([hairyhenderson](https://github.com/hairyhenderson))
- Updating README [\#88](https://github.com/hairyhenderson/gomplate/pull/88) ([rdbaron](https://github.com/rdbaron))

## [v1.2.3](https://github.com/hairyhenderson/gomplate/tree/v1.2.3) (2016-11-24)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.2.2...v1.2.3)

**Fixed bugs:**

- gomplate with vault datasource panics when environment variables are unset [\#83](https://github.com/hairyhenderson/gomplate/issues/83)
- Fixing bug where vault data is incorrectly cached [\#87](https://github.com/hairyhenderson/gomplate/pull/87) ([hairyhenderson](https://github.com/hairyhenderson))
- No vault addr dont panic [\#85](https://github.com/hairyhenderson/gomplate/pull/85) ([drmdrew](https://github.com/drmdrew))

**Merged pull requests:**

- Mention CLI and datasources support in description [\#86](https://github.com/hairyhenderson/gomplate/pull/86) ([drmdrew](https://github.com/drmdrew))

## [v1.2.2](https://github.com/hairyhenderson/gomplate/tree/v1.2.2) (2016-11-20)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.2.1...v1.2.2)

**Implemented enhancements:**

- Adding support for GitHub auth strategy for Vault datasources [\#80](https://github.com/hairyhenderson/gomplate/pull/80) ([hairyhenderson](https://github.com/hairyhenderson))

**Fixed bugs:**

- gomplate w/vault error: user: Current not implemented on linux/amd64  [\#79](https://github.com/hairyhenderson/gomplate/issues/79)
- Avoiding CGO landmine [\#81](https://github.com/hairyhenderson/gomplate/pull/81) ([hairyhenderson](https://github.com/hairyhenderson))

## [v1.2.1](https://github.com/hairyhenderson/gomplate/tree/v1.2.1) (2016-11-19)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.2.0...v1.2.1)

**Fixed bugs:**

- Removing vestigial newline addition [\#77](https://github.com/hairyhenderson/gomplate/pull/77) ([hairyhenderson](https://github.com/hairyhenderson))
- Handle redirects from vault server versions earlier than v0.6.2 [\#76](https://github.com/hairyhenderson/gomplate/pull/76) ([drmdrew](https://github.com/drmdrew))

**Closed issues:**

- Handle vault HTTP redirects [\#75](https://github.com/hairyhenderson/gomplate/issues/75)

## [v1.2.0](https://github.com/hairyhenderson/gomplate/tree/v1.2.0) (2016-11-15)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.1.2...v1.2.0)

**Implemented enhancements:**

- Support for Vault datasources \(app-id & token auth\) [\#74](https://github.com/hairyhenderson/gomplate/pull/74) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding Dockerfile [\#68](https://github.com/hairyhenderson/gomplate/pull/68) ([hairyhenderson](https://github.com/hairyhenderson))

**Merged pull requests:**

- Added a read me section about multiple line if/else/end statements. [\#73](https://github.com/hairyhenderson/gomplate/pull/73) ([EtienneDufresne](https://github.com/EtienneDufresne))
- Adding instructions for installing via the homebrew tap [\#72](https://github.com/hairyhenderson/gomplate/pull/72) ([hairyhenderson](https://github.com/hairyhenderson))
- Updating codegangsta/cli reference to urfave/cli [\#70](https://github.com/hairyhenderson/gomplate/pull/70) ([hairyhenderson](https://github.com/hairyhenderson))
- Formatting with gofmt -s [\#66](https://github.com/hairyhenderson/gomplate/pull/66) ([hairyhenderson](https://github.com/hairyhenderson))

## [v1.1.2](https://github.com/hairyhenderson/gomplate/tree/v1.1.2) (2016-09-06)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.1.1...v1.1.2)

**Fixed bugs:**

- Fixing a panic in Ec2Info.go [\#62](https://github.com/hairyhenderson/gomplate/pull/62) ([marcboudreau](https://github.com/marcboudreau))

## [v1.1.1](https://github.com/hairyhenderson/gomplate/tree/v1.1.1) (2016-09-04)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.1.0...v1.1.1)

**Implemented enhancements:**

- Caching responses from EC2 [\#61](https://github.com/hairyhenderson/gomplate/pull/61) ([hairyhenderson](https://github.com/hairyhenderson))
- Short-circuit ec2 function defaults when not in AWS [\#60](https://github.com/hairyhenderson/gomplate/pull/60) ([hairyhenderson](https://github.com/hairyhenderson))

**Fixed bugs:**

- Slow and repeated network calls during ec2 functions [\#59](https://github.com/hairyhenderson/gomplate/issues/59)

## [v1.1.0](https://github.com/hairyhenderson/gomplate/tree/v1.1.0) (2016-09-02)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.0.0...v1.1.0)

**Implemented enhancements:**

- Provide default when region can't be found [\#55](https://github.com/hairyhenderson/gomplate/issues/55)
- Adding ability to provide default for ec2region function [\#58](https://github.com/hairyhenderson/gomplate/pull/58) ([hairyhenderson](https://github.com/hairyhenderson))

**Merged pull requests:**

- Fixing broken tests [\#57](https://github.com/hairyhenderson/gomplate/pull/57) ([hairyhenderson](https://github.com/hairyhenderson))

## [v1.0.0](https://github.com/hairyhenderson/gomplate/tree/v1.0.0) (2016-07-14)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v0.6.0...v1.0.0)

**Implemented enhancements:**

- Support HTTP/HTTPS datasources [\#45](https://github.com/hairyhenderson/gomplate/issues/45)
- Adding support for HTTP/HTTPS datasources [\#53](https://github.com/hairyhenderson/gomplate/pull/53) ([hairyhenderson](https://github.com/hairyhenderson))

## [v0.6.0](https://github.com/hairyhenderson/gomplate/tree/v0.6.0) (2016-07-12)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v0.5.1...v0.6.0)

**Implemented enhancements:**

- Support YAML data sources [\#43](https://github.com/hairyhenderson/gomplate/issues/43)
- Adding YAML support [\#52](https://github.com/hairyhenderson/gomplate/pull/52) ([hairyhenderson](https://github.com/hairyhenderson))

## [v0.5.1](https://github.com/hairyhenderson/gomplate/tree/v0.5.1) (2016-06-21)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v0.5.0...v0.5.1)

**Fixed bugs:**

- Gomplate sometimes stalls for 5s [\#48](https://github.com/hairyhenderson/gomplate/issues/48)
- Make things start faster [\#51](https://github.com/hairyhenderson/gomplate/pull/51) ([hairyhenderson](https://github.com/hairyhenderson))

## [v0.5.0](https://github.com/hairyhenderson/gomplate/tree/v0.5.0) (2016-05-22)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v0.4.0...v0.5.0)

**Implemented enhancements:**

- It'd be nice to also resolve templates from files [\#8](https://github.com/hairyhenderson/gomplate/issues/8)
- Switching argument parsing to codegangsta/cli [\#42](https://github.com/hairyhenderson/gomplate/pull/42) ([hairyhenderson](https://github.com/hairyhenderson))
- New datasource function - allows use of JSON files as a data source for the template [\#9](https://github.com/hairyhenderson/gomplate/pull/9) ([hairyhenderson](https://github.com/hairyhenderson))

**Fixed bugs:**

- Fixing broken versions in build-x target [\#38](https://github.com/hairyhenderson/gomplate/pull/38) ([hairyhenderson](https://github.com/hairyhenderson))

## [v0.4.0](https://github.com/hairyhenderson/gomplate/tree/v0.4.0) (2016-04-12)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v0.3.0...v0.4.0)

**Implemented enhancements:**

- New functions join, title, toLower, and toUpper [\#36](https://github.com/hairyhenderson/gomplate/pull/36) ([hairyhenderson](https://github.com/hairyhenderson))

## [v0.3.0](https://github.com/hairyhenderson/gomplate/tree/v0.3.0) (2016-04-11)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v0.2.2...v0.3.0)

**Implemented enhancements:**

- Adding slice and jsonArray template functions [\#34](https://github.com/hairyhenderson/gomplate/pull/34) ([hairyhenderson](https://github.com/hairyhenderson))

**Closed issues:**

- gomplate -v returns 0.1.0 even for newer releases [\#33](https://github.com/hairyhenderson/gomplate/issues/33)

**Merged pull requests:**

- Setting the version at build time from the latest tag [\#35](https://github.com/hairyhenderson/gomplate/pull/35) ([hairyhenderson](https://github.com/hairyhenderson))

## [v0.2.2](https://github.com/hairyhenderson/gomplate/tree/v0.2.2) (2016-03-28)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v0.2.1...v0.2.2)

**Fixed bugs:**

- Fixing -v flag [\#32](https://github.com/hairyhenderson/gomplate/pull/32) ([hairyhenderson](https://github.com/hairyhenderson))

## [v0.2.1](https://github.com/hairyhenderson/gomplate/tree/v0.2.1) (2016-03-28)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v0.2.0...v0.2.1)

**Fixed bugs:**

- AWS-dependent functions should fail gracefully when not running in AWS [\#26](https://github.com/hairyhenderson/gomplate/issues/26)
- It's 'ec2region', not 'region' [\#29](https://github.com/hairyhenderson/gomplate/pull/29) ([hairyhenderson](https://github.com/hairyhenderson))
- Using defaults on network errors and timeouts [\#27](https://github.com/hairyhenderson/gomplate/pull/27) ([hairyhenderson](https://github.com/hairyhenderson))

## [v0.2.0](https://github.com/hairyhenderson/gomplate/tree/v0.2.0) (2016-03-28)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v0.1.1...v0.2.0)

**Implemented enhancements:**

- Would be nifty to be able to resolve EC2 metadata [\#15](https://github.com/hairyhenderson/gomplate/issues/15)
- Adding ec2tag and ec2region functions [\#24](https://github.com/hairyhenderson/gomplate/pull/24) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding new ec2dynamic function [\#23](https://github.com/hairyhenderson/gomplate/pull/23) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding json filter function [\#21](https://github.com/hairyhenderson/gomplate/pull/21) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding ec2meta function [\#20](https://github.com/hairyhenderson/gomplate/pull/20) ([hairyhenderson](https://github.com/hairyhenderson))

**Merged pull requests:**

- 📖 Documenting the ec2meta function [\#22](https://github.com/hairyhenderson/gomplate/pull/22) ([hairyhenderson](https://github.com/hairyhenderson))
- 💄 Refactoring to split functions [\#19](https://github.com/hairyhenderson/gomplate/pull/19) ([hairyhenderson](https://github.com/hairyhenderson))

## [v0.1.1](https://github.com/hairyhenderson/gomplate/tree/v0.1.1) (2016-03-22)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v0.1.0...v0.1.1)

**Implemented enhancements:**

- Should support default values for environment variables [\#14](https://github.com/hairyhenderson/gomplate/issues/14)

**Merged pull requests:**

- Updating README with docs for getenv with default [\#17](https://github.com/hairyhenderson/gomplate/pull/17) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding support to getenv for a default value [\#16](https://github.com/hairyhenderson/gomplate/pull/16) ([hairyhenderson](https://github.com/hairyhenderson))

## [v0.1.0](https://github.com/hairyhenderson/gomplate/tree/v0.1.0) (2016-02-20)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v0.0.2...v0.1.0)

**Merged pull requests:**

- Adding new functions `bool` and `getenv` [\#10](https://github.com/hairyhenderson/gomplate/pull/10) ([hairyhenderson](https://github.com/hairyhenderson))
- 📖 Adding details to README [\#7](https://github.com/hairyhenderson/gomplate/pull/7) ([hairyhenderson](https://github.com/hairyhenderson))

## [v0.0.2](https://github.com/hairyhenderson/gomplate/tree/v0.0.2) (2016-01-24)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v0.0.1...v0.0.2)

**Merged pull requests:**

- 💄 slight refactoring & adding some vague unit tests... [\#5](https://github.com/hairyhenderson/gomplate/pull/5) ([hairyhenderson](https://github.com/hairyhenderson))
- 💄 slight refactoring & adding some vague unit tests... [\#4](https://github.com/hairyhenderson/gomplate/pull/4) ([hairyhenderson](https://github.com/hairyhenderson))

## [v0.0.1](https://github.com/hairyhenderson/gomplate/tree/v0.0.1) (2016-01-23)


\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*