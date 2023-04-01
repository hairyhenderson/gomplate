module github.com/hairyhenderson/gomplate/v4

go 1.20

require (
	github.com/Masterminds/goutils v1.1.1
	github.com/Masterminds/semver v1.5.0
	github.com/Shopify/ejson v1.3.3
	github.com/aws/aws-sdk-go v1.44.234
	github.com/docker/libkv v0.2.2-0.20180912205406-458977154600
	github.com/fullsailor/pkcs7 v0.0.0-20190404230743-d7302db945fa
	github.com/go-git/go-billy/v5 v5.4.1
	github.com/go-git/go-git/v5 v5.6.1
	github.com/google/uuid v1.3.0
	github.com/gosimple/slug v1.13.1
	github.com/hairyhenderson/go-fsimpl v0.0.0-20230401192735-3ba7d0a6726a
	github.com/hairyhenderson/toml v0.4.2-0.20210923231440-40456b8e66cf
	github.com/hashicorp/consul/api v1.20.0
	github.com/hashicorp/go-sockaddr v1.0.2
	github.com/hashicorp/vault/api v1.9.0
	github.com/itchyny/gojq v0.12.12
	github.com/johannesboyne/gofakes3 v0.0.0-20230310080033-c0edf658332b
	github.com/joho/godotenv v1.5.1
	github.com/rs/zerolog v1.29.0
	github.com/spf13/afero v1.9.5
	github.com/spf13/cobra v1.6.1
	github.com/stretchr/testify v1.8.2
	github.com/ugorji/go/codec v1.2.11
	github.com/zealic/xignore v0.3.3
	go4.org/netipx v0.0.0-20230303233057-f1b76eb4bb35
	gocloud.dev v0.29.0
	golang.org/x/crypto v0.7.0
	golang.org/x/sys v0.6.0
	golang.org/x/term v0.6.0
	golang.org/x/text v0.8.0
	gotest.tools/v3 v3.4.0
	inet.af/netaddr v0.0.0-20220811202034-502d2d690317
	k8s.io/client-go v0.26.3
)

// TODO: replace with gopkg.in/yaml.v3 after https://github.com/go-yaml/yaml/pull/862
// is merged
require github.com/hairyhenderson/yaml v0.0.0-20220618171115-2d35fca545ce

require (
	cloud.google.com/go v0.110.0 // indirect
	cloud.google.com/go/compute v1.19.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/iam v0.13.0 // indirect
	cloud.google.com/go/storage v1.30.1 // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230331115716-d34776aa93ec // indirect
	github.com/acomagu/bufpipe v1.0.4 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.17.7 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.18.19 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.18 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.60 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.31 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.25 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.32 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.26 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.25 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.14.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.31.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.18.7 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/cloudflare/circl v1.3.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/gojson v0.0.0-20160307161227-2e71ec9dd5ad // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/wire v0.5.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.8.0 // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.2 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.7 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/golang-lru v0.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/imdario/mergo v0.3.15 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/itchyny/timefmt-go v0.1.5 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/ryszard/goskiplist v0.0.0-20150312221310-2dfbae5fcf46 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/shabbyrobe/gocovmerge v0.0.0-20190829150210-3e036491d500 // indirect
	github.com/skeema/knownhosts v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	go.opencensus.io v0.24.0 // indirect
	go4.org/intern v0.0.0-20230205224052-192e9f60865c // indirect
	go4.org/unsafe/assume-no-moving-gc v0.0.0-20230221090011-e4bae7ad2296 // indirect
	golang.org/x/mod v0.9.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/oauth2 v0.6.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.7.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/api v0.114.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230331144136-dcfb400f0633 // indirect
	google.golang.org/grpc v1.54.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
