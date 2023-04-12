package funcs

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"github.com/hairyhenderson/gomplate/v4/azure"
)

// AzureNS - the azure namespace
//
// Deprecated: don't use
func AzureNS() *AzureFuncs {
	return &AzureFuncs{azureOpts: azure.GetClientOptions()}
}

// AddAzureFuncs -
//
// Deprecated: use [CreateAzureFuncs] instead
func AddAzureFuncs(f map[string]interface{}) {
	for k, v := range CreateAzureFuncs(context.Background()) {
		f[k] = v
	}
}

// CreateAzureFuncs -
func CreateAzureFuncs(ctx context.Context) map[string]interface{} {
	ns := &AzureFuncs{
		ctx:       ctx,
		azureOpts: azure.GetClientOptions(),
	}
	return map[string]interface{}{
		"azure": func() interface{} { return ns },
	}
}

// AzureFuncs -
type AzureFuncs struct {
	ctx context.Context

	meta      *azure.MetaClient
	metaInit  sync.Once
	azureOpts azure.ClientOptions
}

// Meta -
func (a *AzureFuncs) Meta(args ...string) (string, error) {
	if len(args) == 0 || len(args) > 3 {
		return "", fmt.Errorf("wrong number of args: wanted 1, 2 or 3, got %d", len(args))
	}
	key := args[0]
	format := "text"
	def := args[1:]
	apiVersion := "2021-12-13"

	if len(args) >= 2 {
		if args[1] == "json" || args[1] == "text" {
			format = args[1]
			def = args[2:]
		}
	}
	if len(args) >= 3 {
		if found, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, args[2]); found {
			apiVersion = args[2]
			def = args[3:]
		}
	}

	a.metaInit.Do(a.initAzureMeta)
	return a.meta.Meta(key, format, apiVersion, def...)
}

func (a *AzureFuncs) initAzureMeta() {
	if a.meta == nil {
		a.meta = azure.NewMetaClient(a.azureOpts)
	}
}
