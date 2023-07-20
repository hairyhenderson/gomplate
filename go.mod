module github.com/flanksource/gomplate/v3

go 1.19

require (
	github.com/Masterminds/goutils v1.1.1
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/flanksource/is-healthy v0.0.0-20230705092916-3b4cf510c5fc
	github.com/google/cel-go v0.17.1
	github.com/google/uuid v1.3.0
	github.com/gosimple/slug v1.13.1
	github.com/hairyhenderson/toml v0.4.2-0.20210923231440-40456b8e66cf
	github.com/itchyny/gojq v0.12.13
	github.com/pkg/errors v0.9.1
	github.com/robertkrimen/otto v0.2.1
	github.com/stretchr/testify v1.8.4
	github.com/ugorji/go/codec v1.2.11
	golang.org/x/text v0.11.0
	golang.org/x/tools v0.7.0
	gotest.tools/v3 v3.4.0
	k8s.io/apimachinery v0.26.4
	sigs.k8s.io/yaml v1.3.0
)

// TODO: replace with gopkg.in/yaml.v3 after https://github.com/go-yaml/yaml/pull/862
// is merged
require github.com/hairyhenderson/yaml v0.0.0-20220618171115-2d35fca545ce

require (
	github.com/antlr/antlr4/runtime/Go/antlr/v4 v4.0.0-20230305170008-8188dc5388df // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/itchyny/timefmt-go v0.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/yuin/gopher-lua v1.1.0 // indirect
	golang.org/x/exp v0.0.0-20220722155223-a9213eeb770e // indirect
	golang.org/x/mod v0.9.0 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230525234035-dd9d682886f9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.26.4 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/utils v0.0.0-20230711102312-30195339c3c7 // indirect
	layeh.com/gopher-json v0.0.0-20201124131017-552bb3c4c3bf // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)
