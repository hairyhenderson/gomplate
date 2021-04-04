package data

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/hairyhenderson/gomplate/v3/internal/datasources"
	"github.com/hairyhenderson/gomplate/v3/vault"
)

func regExtension(ext, typ string) {
	err := mime.AddExtensionType(ext, typ)
	if err != nil {
		panic(err)
	}
}

func init() {
	// Add some types we want to be able to handle which can be missing by default
	regExtension(".json", jsonMimetype)
	regExtension(".yml", yamlMimetype)
	regExtension(".yaml", yamlMimetype)
	regExtension(".csv", csvMimetype)
	regExtension(".toml", tomlMimetype)
	regExtension(".env", envMimetype)
}

// Data -
type Data struct {
	ctx context.Context

	reg datasources.Registry

	// cache map[string][]byte

	// headers from the --datasource-header/-H option that don't reference datasources from the commandline
	extraHeaders map[string]http.Header
}

// Cleanup - clean up datasources before shutting the process down - things
// like Logging out happen here
func (d *Data) Cleanup() {
	// TODO:

	// for _, s := range d.ds {
	// 	s.cleanup()
	// }
}

// NewData - constructor for Data
// Deprecated: will be replaced in future
func NewData(datasourceArgs, headerArgs []string) (*Data, error) {
	cfg := &config.Config{}
	err := cfg.ParseDataSourceFlags(datasourceArgs, nil)
	if err != nil {
		return nil, err
	}
	err = cfg.ParseHeaderFlags(headerArgs)
	if err != nil {
		return nil, err
	}

	data := FromConfig(context.Background(), cfg)

	return data, nil
}

// FromConfig - internal use only!
// Deprecated: will be removed!
func FromConfig(ctx context.Context, cfg *config.Config) *Data {
	// we can't store them in *Data anymore - register directly
	for alias, ds := range cfg.DataSources {
		datasources.DefaultRegistry.Register(alias, ds)
	}

	return &Data{
		ctx:          ctx,
		extraHeaders: cfg.ExtraHeaders,
		reg:          datasources.DefaultRegistry,
	}
}

// Source - a data source
type Source struct {
	URL               *url.URL
	vc                *vault.Vault            // used for vault: URLs, nil otherwise
	asmpg             awssmpGetter            // used for aws+smp:, nil otherwise
	awsSecretsManager awsSecretsManagerGetter // used for aws+sm, nil otherwise
	Alias             string
	mediaType         string
}

// DefineDatasource -
func (d *Data) DefineDatasource(alias, value string) (string, error) {
	if alias == "" {
		return "", errors.New("datasource alias must be provided")
	}
	if d.DatasourceExists(alias) {
		return "", nil
	}
	srcURL, err := config.ParseSourceURL(value)
	if err != nil {
		return "", err
	}
	s := config.DataSource{
		URL:    srcURL,
		Header: d.extraHeaders[alias],
	}
	d.reg.Register(alias, s)

	return "", nil
}

// DatasourceExists -
func (d *Data) DatasourceExists(alias string) bool {
	_, ok := d.reg.Lookup(alias)
	return ok
}

func (d *Data) lookupSource(alias string) (config.DataSource, error) {
	// var ok bool
	ds, ok := d.reg.Lookup(alias)
	if !ok {
		srcURL, err := url.Parse(alias)
		if err != nil || !srcURL.IsAbs() {
			return ds, fmt.Errorf("undefined datasource %q", alias)
		}
		ds.URL = srcURL
		ds.Header = d.extraHeaders[alias]

		d.reg.Register(alias, ds)
	}
	return ds, nil
}

// Include -
func (d *Data) Include(alias string, args ...string) (string, error) {
	ds, err := d.lookupSource(alias)
	if err != nil {
		return "", err
	}

	_, b, err := datasources.ReadDataSource(d.ctx, ds, args...)
	return string(b), err
}

// Datasource -
func (d *Data) Datasource(alias string, args ...string) (interface{}, error) {
	ds, err := d.lookupSource(alias)
	if err != nil {
		return "", err
	}

	resp, err := datasources.Request(d.ctx, ds, args...)
	if err != nil {
		return nil, err
	}

	return resp.Parse()
}

// DatasourceReachable - Determines if the named datasource is reachable with
// the given arguments. Reads from the datasource, and discards the returned data.
func (d *Data) DatasourceReachable(alias string, args ...string) bool {
	ds, ok := d.reg.Lookup(alias)
	if !ok {
		return false
	}

	_, _, err := datasources.ReadDataSource(d.ctx, ds, args...)
	return err == nil
}
