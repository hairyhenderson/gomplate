# Change Log

## [4.1.0](https://github.com/hairyhenderson/gomplate/compare/v4.0.1...v4.1.0) (2024-07-06)


### Features

* **strings:** New functions TrimRight and TrimLeft ([#2148](https://github.com/hairyhenderson/gomplate/issues/2148)) ([bdf3a1e](https://github.com/hairyhenderson/gomplate/commit/bdf3a1eb92020a0d1ce202df14b49f2f13445476))


### Bug Fixes

* **vault:** Upgrade go-fsimpl for KVv2 vault bug, and add test coverage ([#2157](https://github.com/hairyhenderson/gomplate/issues/2157)) ([6ffd703](https://github.com/hairyhenderson/gomplate/commit/6ffd7039b439dbdc40c63b19c85d7f1015ed842d))


### Documentation

* **datasources:** clarify state of Vault KV v2 support ([#2154](https://github.com/hairyhenderson/gomplate/issues/2154)) ([c9643ca](https://github.com/hairyhenderson/gomplate/commit/c9643cad84f95ac0086f8caa0b868364741aa6e6))
* **fix:** Fix broken links, add CI to check ([#2156](https://github.com/hairyhenderson/gomplate/issues/2156)) ([bdf4f8c](https://github.com/hairyhenderson/gomplate/commit/bdf4f8c7d802c6f8ce4bbe6418d583a1449fe493))
* **fix:** Update docs configs to work with the latest hugo theme version ([#2155](https://github.com/hairyhenderson/gomplate/issues/2155)) ([17eb360](https://github.com/hairyhenderson/gomplate/commit/17eb360dfaeaf3186b736971f45f3c418d583845))


### Dependencies

* **actions:** Bump docker/setup-buildx-action from 3.3.0 to 3.4.0 ([#2163](https://github.com/hairyhenderson/gomplate/issues/2163)) ([129ff6b](https://github.com/hairyhenderson/gomplate/commit/129ff6bde8a1fb46b0c2e52586f94cd1b470720b))
* **actions:** Bump docker/setup-qemu-action from 3.0.0 to 3.1.0 ([#2160](https://github.com/hairyhenderson/gomplate/issues/2160)) ([16ebbbe](https://github.com/hairyhenderson/gomplate/commit/16ebbbedf9d6b328c8012933242fbb93b6e3613c))
* **go:** Bump github.com/aws/aws-sdk-go from 1.54.10 to 1.54.11 ([#2152](https://github.com/hairyhenderson/gomplate/issues/2152)) ([e0a6e4f](https://github.com/hairyhenderson/gomplate/commit/e0a6e4f5d707513ef4c33ae8e019da455a7394b6))
* **go:** Bump github.com/aws/aws-sdk-go from 1.54.11 to 1.54.13 ([#2158](https://github.com/hairyhenderson/gomplate/issues/2158)) ([720c70c](https://github.com/hairyhenderson/gomplate/commit/720c70c26b958be784577a349ec2b3a1160e0e54))
* **go:** Bump github.com/aws/aws-sdk-go from 1.54.13 to 1.54.14 ([#2159](https://github.com/hairyhenderson/gomplate/issues/2159)) ([114c54d](https://github.com/hairyhenderson/gomplate/commit/114c54df69738156a70079b5de3352a032c755f9))
* **go:** Bump github.com/aws/aws-sdk-go from 1.54.14 to 1.54.15 ([#2165](https://github.com/hairyhenderson/gomplate/issues/2165)) ([51947a7](https://github.com/hairyhenderson/gomplate/commit/51947a7d5ca7d797ee4998aadfcf856abc8f7a67))
* **go:** Bump github.com/hairyhenderson/go-fsimpl from 0.1.6 to 0.1.7 ([#2167](https://github.com/hairyhenderson/gomplate/issues/2167)) ([80b7c5a](https://github.com/hairyhenderson/gomplate/commit/80b7c5a1aba49239b336d7eeed2525acc2d361be))
* **go:** Bump golang.org/x/term from 0.21.0 to 0.22.0 ([#2162](https://github.com/hairyhenderson/gomplate/issues/2162)) ([59192ec](https://github.com/hairyhenderson/gomplate/commit/59192ec7efe1b59fd800fe399ee5fe063f80287b))

## [4.0.1](https://github.com/hairyhenderson/gomplate/compare/v4.0.0...v4.0.1) (2024-06-28)


### Bug Fixes

* **test:** Fix strings.Repeat test that failed in the wrong way on 32bit ([#2129](https://github.com/hairyhenderson/gomplate/issues/2129)) ([6290186](https://github.com/hairyhenderson/gomplate/commit/62901868f10e887f602e85b37eac70c77f864cc4))


### Documentation

* **chore:** Uncomment 'released' tags for functions in v4.0.0 ([#2125](https://github.com/hairyhenderson/gomplate/issues/2125)) ([e3b86e8](https://github.com/hairyhenderson/gomplate/commit/e3b86e89fca0aad9f5a4f9856f0b57d9cc693470))


### Dependencies

* **go:** Bump cuelang.org/go from 0.9.1 to 0.9.2 ([#2142](https://github.com/hairyhenderson/gomplate/issues/2142)) ([720960e](https://github.com/hairyhenderson/gomplate/commit/720960eb9f25d4d63a037a17648891b8fcf07275))
* **go:** Bump github.com/aws/aws-sdk-go from 1.54.2 to 1.54.6 ([699a2ed](https://github.com/hairyhenderson/gomplate/commit/699a2ed2e202ada74b5c1150a1f6939dff509c86))
* **go:** Bump github.com/aws/aws-sdk-go from 1.54.6 to 1.54.8 ([#2139](https://github.com/hairyhenderson/gomplate/issues/2139)) ([a3475c0](https://github.com/hairyhenderson/gomplate/commit/a3475c01e7afe9b5361dd455434244d6c24f7875))
* **go:** Bump github.com/aws/aws-sdk-go from 1.54.8 to 1.54.9 ([#2143](https://github.com/hairyhenderson/gomplate/issues/2143)) ([715f2c2](https://github.com/hairyhenderson/gomplate/commit/715f2c263f1f6a2c3cb46e4bd5e7996d3587a2e2))
* **go:** Bump github.com/aws/aws-sdk-go from 1.54.9 to 1.54.10 ([#2145](https://github.com/hairyhenderson/gomplate/issues/2145)) ([227b65d](https://github.com/hairyhenderson/gomplate/commit/227b65df1c23504c52428ad49dd42496b353f573))
* **go:** bump github.com/hack-pad/hackpadfs ([#2127](https://github.com/hairyhenderson/gomplate/issues/2127)) ([e6c032b](https://github.com/hairyhenderson/gomplate/commit/e6c032bf458473ff07f0591bef7021e99a851757))
* **go:** Bump github.com/hack-pad/hackpadfs from 0.2.2 to 0.2.3 ([#2131](https://github.com/hairyhenderson/gomplate/issues/2131)) ([4805247](https://github.com/hairyhenderson/gomplate/commit/48052470edcdd5cb3dc8b6ab4ec5bea3048f23a6))
* **go:** Bump github.com/hack-pad/hackpadfs from 0.2.3 to 0.2.4 ([#2137](https://github.com/hairyhenderson/gomplate/issues/2137)) ([eddceaa](https://github.com/hairyhenderson/gomplate/commit/eddceaaf98f0ebd427b154a4bd777c3116112dd6))
* **go:** Bump github.com/hairyhenderson/go-fsimpl from 0.1.4 to 0.1.5 ([#2146](https://github.com/hairyhenderson/gomplate/issues/2146)) ([7e425e1](https://github.com/hairyhenderson/gomplate/commit/7e425e17dbdf561244fa97404f2739bce31b7369))
* **go:** bump github.com/hairyhenderson/go-fsimpl to fix 32-bit panic ([#2128](https://github.com/hairyhenderson/gomplate/issues/2128)) ([5104b19](https://github.com/hairyhenderson/gomplate/commit/5104b19ded072d8ed286cbb41168fb55edb63064))

## [v2.7.0](https://github.com/hairyhenderson/gomplate/tree/v2.7.0) (2018-07-27)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v2.6.0...v2.7.0)

**Implemented enhancements:**

- Adding slice/array support to conv.Has [\#365](https://github.com/hairyhenderson/gomplate/pull/365) ([hairyhenderson](https://github.com/hairyhenderson))
- Allowing datasources to be defined dynamically [\#357](https://github.com/hairyhenderson/gomplate/pull/357) ([hairyhenderson](https://github.com/hairyhenderson))

**Fixed bugs:**

- Require alias for defineDatasource [\#358](https://github.com/hairyhenderson/gomplate/pull/358) ([hairyhenderson](https://github.com/hairyhenderson))

**Closed issues:**

- Feature Request: Allow datasources to be defined dynamically [\#349](https://github.com/hairyhenderson/gomplate/issues/349)
- Can't evaluate field Trunc in type \*funcs.StringFuncs [\#347](https://github.com/hairyhenderson/gomplate/issues/347)

**Merged pull requests:**

- Generating docs [\#366](https://github.com/hairyhenderson/gomplate/pull/366) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding new strings.Sort function [\#364](https://github.com/hairyhenderson/gomplate/pull/364) ([hairyhenderson](https://github.com/hairyhenderson))
- Reducing output on template errors [\#362](https://github.com/hairyhenderson/gomplate/pull/362) ([hairyhenderson](https://github.com/hairyhenderson))
- Move integration tests [\#361](https://github.com/hairyhenderson/gomplate/pull/361) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding fail and assert functions [\#360](https://github.com/hairyhenderson/gomplate/pull/360) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding conv.ToBool/conv.ToBools functions [\#359](https://github.com/hairyhenderson/gomplate/pull/359) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding new defineDatasource function [\#356](https://github.com/hairyhenderson/gomplate/pull/356) ([hairyhenderson](https://github.com/hairyhenderson))
- New path/filepath function namespaces [\#355](https://github.com/hairyhenderson/gomplate/pull/355) ([hairyhenderson](https://github.com/hairyhenderson))
- Add conv.ToStrings function [\#354](https://github.com/hairyhenderson/gomplate/pull/354) ([hairyhenderson](https://github.com/hairyhenderson))
- Bump golang from 1.10-alpine to 1.10.3-alpine [\#353](https://github.com/hairyhenderson/gomplate/pull/353) ([dependabot[bot]](https://github.com/apps/dependabot))
- Bump alpine from 3.7 to 3.8 [\#352](https://github.com/hairyhenderson/gomplate/pull/352) ([dependabot[bot]](https://github.com/apps/dependabot))
- Update golang:1.10-alpine Docker digest to 1c53b8 [\#351](https://github.com/hairyhenderson/gomplate/pull/351) ([renovate[bot]](https://github.com/apps/renovate))
- Update alpine:3.7 Docker digest to 5ce5f5 [\#350](https://github.com/hairyhenderson/gomplate/pull/350) ([renovate[bot]](https://github.com/apps/renovate))
- Update golang:1.10-alpine Docker digest to 79d51d [\#348](https://github.com/hairyhenderson/gomplate/pull/348) ([renovate[bot]](https://github.com/apps/renovate))

## [v2.6.0](https://github.com/hairyhenderson/gomplate/tree/v2.6.0) (2018-06-09)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v2.5.0...v2.6.0)

**Implemented enhancements:**

- Directory datasource [\#215](https://github.com/hairyhenderson/gomplate/issues/215)

**Fixed bugs:**

- The `sockaddr.Include` and `sockaddr.Exclude` do not have "private" selector. [\#328](https://github.com/hairyhenderson/gomplate/issues/328)
- Support commas in number conversion [\#345](https://github.com/hairyhenderson/gomplate/pull/345) ([hairyhenderson](https://github.com/hairyhenderson))

**Closed issues:**

- Add slugify [\#336](https://github.com/hairyhenderson/gomplate/issues/336)
- Action Required: Fix Renovate Configuration [\#335](https://github.com/hairyhenderson/gomplate/issues/335)
- Consider publishing sha256sums of release files [\#318](https://github.com/hairyhenderson/gomplate/issues/318)
- Vault list support [\#229](https://github.com/hairyhenderson/gomplate/issues/229)

**Merged pull requests:**

- Update golang:1.10-alpine Docker digest to 56db23 [\#346](https://github.com/hairyhenderson/gomplate/pull/346) ([renovate[bot]](https://github.com/apps/renovate))
- Update golang:1.10-alpine Docker digest to bb3108 [\#344](https://github.com/hairyhenderson/gomplate/pull/344) ([renovate[bot]](https://github.com/apps/renovate))
- Adding env datasource [\#341](https://github.com/hairyhenderson/gomplate/pull/341) ([hairyhenderson](https://github.com/hairyhenderson))
- Add strings.Slug function [\#339](https://github.com/hairyhenderson/gomplate/pull/339) ([hairyhenderson](https://github.com/hairyhenderson))
- Updating vendored packages [\#338](https://github.com/hairyhenderson/gomplate/pull/338) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding directory support for file datasources [\#334](https://github.com/hairyhenderson/gomplate/pull/334) ([hairyhenderson](https://github.com/hairyhenderson))
- Overhauling datasource documentation [\#333](https://github.com/hairyhenderson/gomplate/pull/333) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding vault list support [\#332](https://github.com/hairyhenderson/gomplate/pull/332) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding more math functions [\#331](https://github.com/hairyhenderson/gomplate/pull/331) ([hairyhenderson](https://github.com/hairyhenderson))
- Add missing anchor for RFC 1918 link in sockaddr documentation. [\#330](https://github.com/hairyhenderson/gomplate/pull/330) ([kwilczynski](https://github.com/kwilczynski))
- Remove notion of "private" selector from the Include/Exclude filter. [\#329](https://github.com/hairyhenderson/gomplate/pull/329) ([kwilczynski](https://github.com/kwilczynski))
- Improving documentation around slim binaries [\#327](https://github.com/hairyhenderson/gomplate/pull/327) ([hairyhenderson](https://github.com/hairyhenderson))
- Update golang:1.10-alpine Docker digest to 96e25c [\#325](https://github.com/hairyhenderson/gomplate/pull/325) ([renovate[bot]](https://github.com/apps/renovate))
- Update alpine:3.7 Docker digest to 8c03bb [\#324](https://github.com/hairyhenderson/gomplate/pull/324) ([renovate[bot]](https://github.com/apps/renovate))
- Adding strings.Trunc and strings.Abbrev [\#323](https://github.com/hairyhenderson/gomplate/pull/323) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding crypto.Bcrypt function [\#322](https://github.com/hairyhenderson/gomplate/pull/322) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding strings.TrimSuffix and strings.Repeat [\#321](https://github.com/hairyhenderson/gomplate/pull/321) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding missing docs for file.Walk and strings.TrimPrefix [\#320](https://github.com/hairyhenderson/gomplate/pull/320) ([hairyhenderson](https://github.com/hairyhenderson))
- Add ability to generate checksums of binaries [\#319](https://github.com/hairyhenderson/gomplate/pull/319) ([hairyhenderson](https://github.com/hairyhenderson))
- Update golang:1.10-alpine Docker digest to 9de80c [\#317](https://github.com/hairyhenderson/gomplate/pull/317) ([renovate[bot]](https://github.com/apps/renovate))

## [v2.5.0](https://github.com/hairyhenderson/gomplate/tree/v2.5.0) (2018-05-04)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v2.4.0...v2.5.0)

**Implemented enhancements:**

- Exec mode... [\#300](https://github.com/hairyhenderson/gomplate/issues/300)
- Need a way to determine whether a datasource is reachable [\#286](https://github.com/hairyhenderson/gomplate/issues/286)
- Add `go-sockaddr` functions [\#145](https://github.com/hairyhenderson/gomplate/issues/145)
- Adding datasourceReachable function [\#315](https://github.com/hairyhenderson/gomplate/pull/315) ([hairyhenderson](https://github.com/hairyhenderson))
- Execute additional command after -- [\#307](https://github.com/hairyhenderson/gomplate/pull/307) ([hairyhenderson](https://github.com/hairyhenderson))

**Fixed bugs:**

- panic when parsing an empty CSV as a datasource [\#311](https://github.com/hairyhenderson/gomplate/issues/311)
- File mode is not preserved [\#296](https://github.com/hairyhenderson/gomplate/issues/296)
- Fixing panic when parsing empty CSVs and CSVs containing only newlines [\#312](https://github.com/hairyhenderson/gomplate/pull/312) ([hairyhenderson](https://github.com/hairyhenderson))
- Avoid closing stdout [\#306](https://github.com/hairyhenderson/gomplate/pull/306) ([hairyhenderson](https://github.com/hairyhenderson))
- Writing output files from a stdin template requires permissions [\#305](https://github.com/hairyhenderson/gomplate/pull/305) ([benjdewan](https://github.com/benjdewan))
- Linting subpackages too... [\#302](https://github.com/hairyhenderson/gomplate/pull/302) ([hairyhenderson](https://github.com/hairyhenderson))

**Closed issues:**

- Writing an output file from a template provided via --in sets no FileMode when run using Docker [\#304](https://github.com/hairyhenderson/gomplate/issues/304)
- function "dict" not defined [\#291](https://github.com/hairyhenderson/gomplate/issues/291)
- unexpected "|" in template clause [\#290](https://github.com/hairyhenderson/gomplate/issues/290)
- Conditional statement as default value in getenv? [\#285](https://github.com/hairyhenderson/gomplate/issues/285)
- Pull in sprig functions? [\#283](https://github.com/hairyhenderson/gomplate/issues/283)
- Consider breaking the gomplate cmd into a sub-package [\#147](https://github.com/hairyhenderson/gomplate/issues/147)

**Merged pull requests:**

- Relaxing restriction on empty datasources [\#316](https://github.com/hairyhenderson/gomplate/pull/316) ([hairyhenderson](https://github.com/hairyhenderson))
- Improving error handling for datasources [\#314](https://github.com/hairyhenderson/gomplate/pull/314) ([hairyhenderson](https://github.com/hairyhenderson))
- Pin alpine Docker tag [\#309](https://github.com/hairyhenderson/gomplate/pull/309) ([renovate[bot]](https://github.com/apps/renovate))
- Adding alpine Docker image variant [\#308](https://github.com/hairyhenderson/gomplate/pull/308) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding basic metrics around template rendering [\#303](https://github.com/hairyhenderson/gomplate/pull/303) ([hairyhenderson](https://github.com/hairyhenderson))
- Preserve FileMode of input file when writing output file [\#301](https://github.com/hairyhenderson/gomplate/pull/301) ([djgilcrease](https://github.com/djgilcrease))
- Exporting the writer used when templates are sent to Stdout [\#299](https://github.com/hairyhenderson/gomplate/pull/299) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding new conv.Default function [\#298](https://github.com/hairyhenderson/gomplate/pull/298) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding time.Since and time.Until funcs [\#295](https://github.com/hairyhenderson/gomplate/pull/295) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding time.ParseDuration function [\#294](https://github.com/hairyhenderson/gomplate/pull/294) ([hairyhenderson](https://github.com/hairyhenderson))
- Relax inputs for many functions [\#293](https://github.com/hairyhenderson/gomplate/pull/293) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding conv.ToString function [\#292](https://github.com/hairyhenderson/gomplate/pull/292) ([hairyhenderson](https://github.com/hairyhenderson))
- chore\(deps\): update golang:1.10-alpine docker digest to 356aea [\#289](https://github.com/hairyhenderson/gomplate/pull/289) ([renovate[bot]](https://github.com/apps/renovate))
- meta: Add release for freebsd-amd64 [\#287](https://github.com/hairyhenderson/gomplate/pull/287) ([jen20](https://github.com/jen20))
- New env.ExpandEnv function [\#284](https://github.com/hairyhenderson/gomplate/pull/284) ([hairyhenderson](https://github.com/hairyhenderson))
- New function proposal: `strings.TrimPrefix` [\#282](https://github.com/hairyhenderson/gomplate/pull/282) ([christopher-avila](https://github.com/christopher-avila))
- New function: `file.Walk` [\#281](https://github.com/hairyhenderson/gomplate/pull/281) ([christopher-avila](https://github.com/christopher-avila))
- Update golang Docker image 1.10-alpine digest \(2d95d3\) [\#280](https://github.com/hairyhenderson/gomplate/pull/280) ([renovate[bot]](https://github.com/apps/renovate))
- Update deps [\#273](https://github.com/hairyhenderson/gomplate/pull/273) ([hairyhenderson](https://github.com/hairyhenderson))
- Putting main pkg in cmd subdirectory [\#264](https://github.com/hairyhenderson/gomplate/pull/264) ([hairyhenderson](https://github.com/hairyhenderson))

## [v2.4.0](https://github.com/hairyhenderson/gomplate/tree/v2.4.0) (2018-03-04)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v2.3.0...v2.4.0)

**Implemented enhancements:**

- Support setting Vault server URL in datasource URL [\#243](https://github.com/hairyhenderson/gomplate/issues/243)
- Exclude option support [\#218](https://github.com/hairyhenderson/gomplate/issues/218)
- Adding sockaddr namespace [\#271](https://github.com/hairyhenderson/gomplate/pull/271) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding file namespace [\#270](https://github.com/hairyhenderson/gomplate/pull/270) ([hairyhenderson](https://github.com/hairyhenderson))

**Fixed bugs:**

- Solaris build is broken 💥 [\#255](https://github.com/hairyhenderson/gomplate/issues/255)
- strings.Indent should not indent when width is 0 [\#268](https://github.com/hairyhenderson/gomplate/pull/268) ([keitwb](https://github.com/keitwb))
- Attempting to fix intermittent Integration Test failure [\#260](https://github.com/hairyhenderson/gomplate/pull/260) ([hairyhenderson](https://github.com/hairyhenderson))

**Closed issues:**

- docker run hairyhenderson/gomplate --version doesn't print version [\#266](https://github.com/hairyhenderson/gomplate/issues/266)

**Merged pull requests:**

- Log test output in CI [\#272](https://github.com/hairyhenderson/gomplate/pull/272) ([hairyhenderson](https://github.com/hairyhenderson))
- Updating vendored dependencies [\#269](https://github.com/hairyhenderson/gomplate/pull/269) ([hairyhenderson](https://github.com/hairyhenderson))
- Fix the example command in 'use with docker' section [\#267](https://github.com/hairyhenderson/gomplate/pull/267) ([yizhiheng](https://github.com/yizhiheng))
- Migrate from bats to pure Go for integration tests [\#265](https://github.com/hairyhenderson/gomplate/pull/265) ([hairyhenderson](https://github.com/hairyhenderson))
- Rebasing Docker images on `scratch` instead of alpine [\#261](https://github.com/hairyhenderson/gomplate/pull/261) ([hairyhenderson](https://github.com/hairyhenderson))
- Building with Go 1.10 [\#258](https://github.com/hairyhenderson/gomplate/pull/258) ([hairyhenderson](https://github.com/hairyhenderson))

## [v2.3.0](https://github.com/hairyhenderson/gomplate/tree/v2.3.0) (2018-02-12)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v2.2.0...v2.3.0)

**Implemented enhancements:**

- Enable reading from AWS SSM Parameter Store? [\#245](https://github.com/hairyhenderson/gomplate/issues/245)
- Can we specify stdin as a datasource? [\#230](https://github.com/hairyhenderson/gomplate/issues/230)
- Trade the dependency on `aws-sdk-go` for something smaller [\#47](https://github.com/hairyhenderson/gomplate/issues/47)
- Allow vault address to be specified in the vault:// URL [\#251](https://github.com/hairyhenderson/gomplate/pull/251) ([hairyhenderson](https://github.com/hairyhenderson))
- Add AWS SSM Parameter support [\#248](https://github.com/hairyhenderson/gomplate/pull/248) ([tyrken](https://github.com/tyrken))
- Add crypto namespace [\#236](https://github.com/hairyhenderson/gomplate/pull/236) ([hairyhenderson](https://github.com/hairyhenderson))
- Support setting MIME type with URL query string [\#234](https://github.com/hairyhenderson/gomplate/pull/234) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding support for stdin: scheme for datasources [\#233](https://github.com/hairyhenderson/gomplate/pull/233) ([hairyhenderson](https://github.com/hairyhenderson))
- Can now pass --exclude as a flag [\#220](https://github.com/hairyhenderson/gomplate/pull/220) ([Gman98ish](https://github.com/Gman98ish))

**Fixed bugs:**

- "unexpected {{end}}" error that only happens when using --input-dir [\#238](https://github.com/hairyhenderson/gomplate/issues/238)

**Closed issues:**

- gomplate should output which template was being parsed when an error is encountered [\#239](https://github.com/hairyhenderson/gomplate/issues/239)
- function "math" not defined [\#224](https://github.com/hairyhenderson/gomplate/issues/224)

**Merged pull requests:**

- new logo [\#253](https://github.com/hairyhenderson/gomplate/pull/253) ([hairyhenderson](https://github.com/hairyhenderson))
- bind test binaries explicitly to 127.0.0.1 [\#252](https://github.com/hairyhenderson/gomplate/pull/252) ([hairyhenderson](https://github.com/hairyhenderson))
- Updating github.com/aws/aws-sdk-go to v1.12.70 [\#247](https://github.com/hairyhenderson/gomplate/pull/247) ([hairyhenderson](https://github.com/hairyhenderson))
- Updating for dep 0.4.0 and revendoring [\#246](https://github.com/hairyhenderson/gomplate/pull/246) ([hairyhenderson](https://github.com/hairyhenderson))
- Increase gometalinter timeout and make it go faster locally [\#244](https://github.com/hairyhenderson/gomplate/pull/244) ([hairyhenderson](https://github.com/hairyhenderson))
- Refactoring template processing [\#241](https://github.com/hairyhenderson/gomplate/pull/241) ([hairyhenderson](https://github.com/hairyhenderson))
- Naming template after input filename [\#240](https://github.com/hairyhenderson/gomplate/pull/240) ([hairyhenderson](https://github.com/hairyhenderson))
- Pruning dependencies with `dep prune` [\#237](https://github.com/hairyhenderson/gomplate/pull/237) ([hairyhenderson](https://github.com/hairyhenderson))

## [v2.2.0](https://github.com/hairyhenderson/gomplate/tree/v2.2.0) (2017-11-03)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v2.1.0...v2.2.0)

**Implemented enhancements:**

- Add some time-related functions [\#199](https://github.com/hairyhenderson/gomplate/issues/199)
- Adding math.Seq function [\#227](https://github.com/hairyhenderson/gomplate/pull/227) ([hairyhenderson](https://github.com/hairyhenderson))
- Add time.ParseLocal and time.ParseInLocation functions [\#223](https://github.com/hairyhenderson/gomplate/pull/223) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding time.ZoneOffset function [\#222](https://github.com/hairyhenderson/gomplate/pull/222) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding conv.ToInt64, conv.ToFloat64, and others... [\#216](https://github.com/hairyhenderson/gomplate/pull/216) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding math functions [\#214](https://github.com/hairyhenderson/gomplate/pull/214) ([hairyhenderson](https://github.com/hairyhenderson))

**Fixed bugs:**

- Fixing conv.Join to handle non-interface{} arrays [\#226](https://github.com/hairyhenderson/gomplate/pull/226) ([hairyhenderson](https://github.com/hairyhenderson))
- Fixing bugs in ToInt64/ToFloat64 [\#217](https://github.com/hairyhenderson/gomplate/pull/217) ([hairyhenderson](https://github.com/hairyhenderson))

**Merged pull requests:**

- Using Go 1.9.x now - go test ignores vendor on its own now [\#228](https://github.com/hairyhenderson/gomplate/pull/228) ([hairyhenderson](https://github.com/hairyhenderson))
- Stabilizing integration tests a bit [\#221](https://github.com/hairyhenderson/gomplate/pull/221) ([hairyhenderson](https://github.com/hairyhenderson))
- Don't panic on template errors [\#219](https://github.com/hairyhenderson/gomplate/pull/219) ([hairyhenderson](https://github.com/hairyhenderson))

## [v2.1.0](https://github.com/hairyhenderson/gomplate/tree/v2.1.0) (2017-10-14)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v2.0.1...v2.1.0)

**Implemented enhancements:**

- Add time funcs [\#211](https://github.com/hairyhenderson/gomplate/pull/211) ([hairyhenderson](https://github.com/hairyhenderson))

**Fixed bugs:**

- 'client nonce mismatch' when using AWS auth with nonce [\#205](https://github.com/hairyhenderson/gomplate/issues/205)
- AWS Auth nonce file not created if it doesn't already exist [\#204](https://github.com/hairyhenderson/gomplate/issues/204)
- "gomplate -in flibbit" should produce an error [\#192](https://github.com/hairyhenderson/gomplate/issues/192)
- Fixes \#192 - fail with unknown args [\#208](https://github.com/hairyhenderson/gomplate/pull/208) ([drmdrew](https://github.com/drmdrew))
- Remove trailing spaces [\#207](https://github.com/hairyhenderson/gomplate/pull/207) ([stuart-c](https://github.com/stuart-c))
- Create file if it doesn't exist [\#206](https://github.com/hairyhenderson/gomplate/pull/206) ([stuart-c](https://github.com/stuart-c))

**Closed issues:**

- Document 4 new conv functions in 2.0.0 [\#196](https://github.com/hairyhenderson/gomplate/issues/196)

**Merged pull requests:**

- Document conv.ParseInt, conv.ParseFloat, conv.ParseUint, and conv.Atoi [\#212](https://github.com/hairyhenderson/gomplate/pull/212) ([danedmunds](https://github.com/danedmunds))

## [v2.0.1](https://github.com/hairyhenderson/gomplate/tree/v2.0.1) (2017-09-08)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v2.0.0...v2.0.1)

**Fixed bugs:**

- Crash when vault secret is not found [\#200](https://github.com/hairyhenderson/gomplate/issues/200)
- Fixing crash on 404 [\#201](https://github.com/hairyhenderson/gomplate/pull/201) ([hairyhenderson](https://github.com/hairyhenderson))

**Merged pull requests:**

- Add nonce support [\#202](https://github.com/hairyhenderson/gomplate/pull/202) ([stuart-c](https://github.com/stuart-c))

## [v2.0.0](https://github.com/hairyhenderson/gomplate/tree/v2.0.0) (2017-08-10)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.10.0...v2.0.0)

**Implemented enhancements:**

- Consul support [\#173](https://github.com/hairyhenderson/gomplate/issues/173)
- Extracting data namespace, renaming typeconv to conv namespace [\#194](https://github.com/hairyhenderson/gomplate/pull/194) ([hairyhenderson](https://github.com/hairyhenderson))
- Vault AWS EC2 auth [\#190](https://github.com/hairyhenderson/gomplate/pull/190) ([stuart-c](https://github.com/stuart-c))
- Consul vault auth [\#187](https://github.com/hairyhenderson/gomplate/pull/187) ([stuart-c](https://github.com/stuart-c))
- Vault write support [\#183](https://github.com/hairyhenderson/gomplate/pull/183) ([stuart-c](https://github.com/stuart-c))
- Add Consul & BoltDB datasource support [\#178](https://github.com/hairyhenderson/gomplate/pull/178) ([stuart-c](https://github.com/stuart-c))

**Closed issues:**

- gomplate --version: 0.0.0 [\#188](https://github.com/hairyhenderson/gomplate/issues/188)

**Merged pull requests:**

- Adding a couple extra integration tests for vault [\#195](https://github.com/hairyhenderson/gomplate/pull/195) ([hairyhenderson](https://github.com/hairyhenderson))
- Moving mustParse functions into new typeconv package [\#193](https://github.com/hairyhenderson/gomplate/pull/193) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding unit tests for libkv datasources [\#189](https://github.com/hairyhenderson/gomplate/pull/189) ([hairyhenderson](https://github.com/hairyhenderson))
- Fixing up typos and formatting in docs [\#186](https://github.com/hairyhenderson/gomplate/pull/186) ([hairyhenderson](https://github.com/hairyhenderson))
- Migrate from glide to dep [\#185](https://github.com/hairyhenderson/gomplate/pull/185) ([hairyhenderson](https://github.com/hairyhenderson))
- Updating dependencies [\#184](https://github.com/hairyhenderson/gomplate/pull/184) ([hairyhenderson](https://github.com/hairyhenderson))
- Switch to using official Go Vault client [\#177](https://github.com/hairyhenderson/gomplate/pull/177) ([stuart-c](https://github.com/stuart-c))

## [v1.10.0](https://github.com/hairyhenderson/gomplate/tree/v1.10.0) (2017-08-01)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.9.1...v1.10.0)

**Implemented enhancements:**

- Adding support for \_FILE fallback to env.Getenv function [\#181](https://github.com/hairyhenderson/gomplate/pull/181) ([hairyhenderson](https://github.com/hairyhenderson))

**Closed issues:**

- 17.7M on Alpine Images [\#171](https://github.com/hairyhenderson/gomplate/issues/171)

**Merged pull requests:**

- Moving getenv to separate package [\#179](https://github.com/hairyhenderson/gomplate/pull/179) ([hairyhenderson](https://github.com/hairyhenderson))
- Remove VFS argument from ReadSource which isn't used [\#175](https://github.com/hairyhenderson/gomplate/pull/175) ([stuart-c](https://github.com/stuart-c))
- Disabling cgo so the binary is truly static [\#174](https://github.com/hairyhenderson/gomplate/pull/174) ([hairyhenderson](https://github.com/hairyhenderson))

## [v1.9.1](https://github.com/hairyhenderson/gomplate/tree/v1.9.1) (2017-06-22)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.9.0...v1.9.1)

**Implemented enhancements:**

- Enhance the indent function [\#165](https://github.com/hairyhenderson/gomplate/issues/165)

**Fixed bugs:**

- gomplate v1.9.0  - fails for aws.EC2 calls that take 1s or plus \(Windows\) [\#168](https://github.com/hairyhenderson/gomplate/issues/168)
- Adding AWS\_TIMEOUT environment variable [\#169](https://github.com/hairyhenderson/gomplate/pull/169) ([hairyhenderson](https://github.com/hairyhenderson))

**Closed issues:**

- Gomplate function to output a gomplate function [\#167](https://github.com/hairyhenderson/gomplate/issues/167)

## [v1.9.0](https://github.com/hairyhenderson/gomplate/tree/v1.9.0) (2017-06-14)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.8.0...v1.9.0)

**Implemented enhancements:**

- DNS resolver function... [\#157](https://github.com/hairyhenderson/gomplate/issues/157)
- Regular expression support [\#152](https://github.com/hairyhenderson/gomplate/issues/152)
- Enhancing indent function [\#166](https://github.com/hairyhenderson/gomplate/pull/166) ([hairyhenderson](https://github.com/hairyhenderson))
- Creating a strings namespace [\#164](https://github.com/hairyhenderson/gomplate/pull/164) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding regexp support [\#161](https://github.com/hairyhenderson/gomplate/pull/161) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding net.Lookup\* functions [\#158](https://github.com/hairyhenderson/gomplate/pull/158) ([hairyhenderson](https://github.com/hairyhenderson))

**Closed issues:**

- JSON formatting [\#163](https://github.com/hairyhenderson/gomplate/issues/163)
- panic: template: template:19:25: executing "template" at \<.Env\>: map has no entry for key "Env" [\#160](https://github.com/hairyhenderson/gomplate/issues/160)
- Suggestion: add directory support for loading environment [\#159](https://github.com/hairyhenderson/gomplate/issues/159)

## [v1.8.0](https://github.com/hairyhenderson/gomplate/tree/v1.8.0) (2017-06-09)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.7.0...v1.8.0)

**Implemented enhancements:**

- base64 encode/decode support [\#155](https://github.com/hairyhenderson/gomplate/issues/155)
- Ability to include raw text from non-structured files [\#142](https://github.com/hairyhenderson/gomplate/issues/142)
- Support CSV datasources  [\#44](https://github.com/hairyhenderson/gomplate/issues/44)
- Adding new base64.Encode/base64.Decode functions [\#156](https://github.com/hairyhenderson/gomplate/pull/156) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding TOML support [\#154](https://github.com/hairyhenderson/gomplate/pull/154) ([hairyhenderson](https://github.com/hairyhenderson))
- Add include function [\#153](https://github.com/hairyhenderson/gomplate/pull/153) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding CSV datasource support [\#150](https://github.com/hairyhenderson/gomplate/pull/150) ([hairyhenderson](https://github.com/hairyhenderson))

**Closed issues:**

- Improve the docs and move to a separate place \(not the README\) [\#146](https://github.com/hairyhenderson/gomplate/issues/146)

**Merged pull requests:**

- Moving docs out of README [\#149](https://github.com/hairyhenderson/gomplate/pull/149) ([hairyhenderson](https://github.com/hairyhenderson))
- Namespacing the aws funcs [\#148](https://github.com/hairyhenderson/gomplate/pull/148) ([hairyhenderson](https://github.com/hairyhenderson))

## [v1.7.0](https://github.com/hairyhenderson/gomplate/tree/v1.7.0) (2017-05-24)
[Full Changelog](https://github.com/hairyhenderson/gomplate/compare/v1.6.0...v1.7.0)

**Implemented enhancements:**

- Add "replace" function and documentation [\#140](https://github.com/hairyhenderson/gomplate/pull/140) ([jen20](https://github.com/jen20))
- Adding new indent function [\#139](https://github.com/hairyhenderson/gomplate/pull/139) ([hairyhenderson](https://github.com/hairyhenderson))
- Adding new toJSONPretty function [\#137](https://github.com/hairyhenderson/gomplate/pull/137) ([hairyhenderson](https://github.com/hairyhenderson))
- Add urlParse function \(i.e. url.Parse\) [\#132](https://github.com/hairyhenderson/gomplate/pull/132) ([hairyhenderson](https://github.com/hairyhenderson))
- Add splitN function \(i.e. strings.SplitN\) [\#131](https://github.com/hairyhenderson/gomplate/pull/131) ([hairyhenderson](https://github.com/hairyhenderson))

**Fixed bugs:**

- toJSON fails to marshal objects with nested objects [\#138](https://github.com/hairyhenderson/gomplate/issues/138)
- function "has" will panic when used on nested map [\#134](https://github.com/hairyhenderson/gomplate/issues/134)
- Using  github.com/ugorji/go/codec for JSON encoding instead of encoding/json [\#144](https://github.com/hairyhenderson/gomplate/pull/144) ([hairyhenderson](https://github.com/hairyhenderson))
- Fixing bug with 'has' and 'datasource' around referencing sub-maps in nested maps [\#135](https://github.com/hairyhenderson/gomplate/pull/135) ([hairyhenderson](https://github.com/hairyhenderson))

**Closed issues:**

- Ability to join list of items into string with separator [\#143](https://github.com/hairyhenderson/gomplate/issues/143)

**Merged pull requests:**

- Add `solaris-amd64` build target [\#141](https://github.com/hairyhenderson/gomplate/pull/141) ([jen20](https://github.com/jen20))
- Making the built Docker image smaller [\#136](https://github.com/hairyhenderson/gomplate/pull/136) ([hairyhenderson](https://github.com/hairyhenderson))

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
