module github.com/flanksource/gomplate/v3

go 1.20

require (
	github.com/Masterminds/goutils v1.1.1
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/flanksource/is-healthy v0.0.0-20231003215854-76c51e3a3ff7
	github.com/flanksource/kubectl-neat v1.0.4
	github.com/google/cel-go v0.18.2
	github.com/google/go-cmp v0.6.0
	github.com/google/uuid v1.5.0
	github.com/gosimple/slug v1.13.1
	github.com/hairyhenderson/toml v0.4.2-0.20210923231440-40456b8e66cf
	github.com/itchyny/gojq v0.12.14
	github.com/mitchellh/reflectwalk v1.0.2
	github.com/ohler55/ojg v1.20.2
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/robertkrimen/otto v0.2.1
	github.com/stretchr/testify v1.8.4
	github.com/ugorji/go/codec v1.2.12
	golang.org/x/text v0.14.0
	golang.org/x/tools v0.16.1
	google.golang.org/protobuf v1.32.0
	gotest.tools/v3 v3.5.1
	k8s.io/api v0.28.2
	k8s.io/apimachinery v0.28.2
	sigs.k8s.io/yaml v1.3.0
)

// TODO: replace with gopkg.in/yaml.v3 after https://github.com/go-yaml/yaml/pull/862
// is merged
require github.com/hairyhenderson/yaml v0.0.0-20220618171115-2d35fca545ce

require (
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/distribution/reference v0.5.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/itchyny/timefmt-go v0.1.5 // indirect
	github.com/jeremywohl/flatten v0.0.0-20180923035001-588fe0d4c603 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/spf13/cobra v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/tidwall/gjson v1.9.3 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tidwall/sjson v1.0.4 // indirect
	github.com/yuin/gopher-lua v1.1.0 // indirect
	golang.org/x/exp v0.0.0-20231005195138-3e424a577f31 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231002182017-d307bd883b97 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231002182017-d307bd883b97 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/client-go v0.28.2 // indirect
	k8s.io/klog/v2 v2.110.1 // indirect
	k8s.io/utils v0.0.0-20240102154912-e7106e64919e // indirect
	layeh.com/gopher-json v0.0.0-20201124131017-552bb3c4c3bf // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
)
