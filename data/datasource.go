package data

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/hairyhenderson/gomplate/v3/internal/datasource"
)

// Data -
type Data struct {
	sourceReg datasource.SourceRegistry
	// headers from the --datasource-header/-H option that don't reference datasources from the commandline
	extraHeaders map[string]http.Header
}

// Cleanup - clean up datasources before shutting the process down - things
// like Logging out happen here
func (d *Data) Cleanup() {
	// for _, s := range d.Sources {
	// 	s.Cleanup()
	// }
}

// NewData - constructor for Data
// Deprecated: use datasource package
func NewData(datasourceArgs, headerArgs []string) (*Data, error) {
	cfg := &config.Config{}
	err := cfg.ParseDataSourceFlags(datasourceArgs, nil, headerArgs)
	data, err := FromConfig(cfg)
	return data, err
}

var sourceRegistry = datasource.DefaultRegistry

// FromConfig - internal use only!
func FromConfig(cfg *config.Config) (data *Data, err error) {
	data = &Data{sourceReg: sourceRegistry}
	for alias, d := range cfg.DataSources {
		_, err = data.sourceReg.Register(alias, d.URL, d.Header)
	}
	for alias, d := range cfg.Context {
		_, err = data.sourceReg.Register(alias, d.URL, d.Header)
	}
	data.extraHeaders = cfg.ExtraHeaders
	return data, err
}

// Source - a data source
type Source struct {
	Alias string
	URL   *url.URL
}

func (s *Source) cleanup() {
}

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (s *Source) String() string {
	return fmt.Sprintf("%s=%s", s.Alias, s.URL.String())
}

func parseSourceURL(value string) (*url.URL, error) {
	if value == "-" {
		value = "stdin://"
	}
	value = filepath.ToSlash(value)
	// handle absolute Windows paths
	volName := ""
	if volName = filepath.VolumeName(value); volName != "" {
		// handle UNCs
		if len(volName) > 2 {
			value = "file:" + value
		} else {
			value = "file:///" + value
		}
	}
	srcURL, err := url.Parse(value)
	if err != nil {
		return nil, err
	}

	if volName != "" {
		if strings.HasPrefix(srcURL.Path, "/") && srcURL.Path[2] == ':' {
			srcURL.Path = srcURL.Path[1:]
		}
	}

	if !srcURL.IsAbs() {
		srcURL, err = absFileURL(value)
		if err != nil {
			return nil, err
		}
	}
	return srcURL, nil
}

func absFileURL(value string) (*url.URL, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrapf(err, "can't get working directory")
	}
	urlCwd := filepath.ToSlash(cwd)
	baseURL := &url.URL{
		Scheme: "file",
		Path:   urlCwd + "/",
	}
	relURL, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("can't parse value %s as URL: %w", value, err)
	}
	resolved := baseURL.ResolveReference(relURL)
	// deal with Windows drive letters
	if !strings.HasPrefix(urlCwd, "/") && resolved.Path[2] == ':' {
		resolved.Path = resolved.Path[1:]
	}
	return resolved, nil
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
	_, err = d.sourceReg.Register(alias, srcURL, d.extraHeaders[alias])
	return "", err
}

// DatasourceExists -
func (d *Data) DatasourceExists(alias string) bool {
	return d.sourceReg.Exists(alias)
}

func (d *Data) lookupSource(alias string) (source datasource.Source, err error) {
	source = d.sourceReg.Get(alias)
	if source == nil {
		source, err = d.sourceReg.Dynamic(alias, d.extraHeaders[alias])
	}
	return source, err
}

func (d *Data) readDataSource(ctx context.Context, alias string, args ...string) (data *datasource.Data, err error) {
	source, err := d.lookupSource(alias)
	if err != nil {
		return nil, err
	}
	data, err = source.Read(ctx, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "Couldn't read datasource '%s'", alias)
	}
	return data, nil
}

// Include -
func (d *Data) Include(alias string, args ...string) (string, error) {
	ctx := context.TODO()
	data, err := d.readDataSource(ctx, alias, args...)
	if err != nil {
		return "", err
	}
	return string(data.Bytes), nil
}

// Datasource -
func (d *Data) Datasource(alias string, args ...string) (interface{}, error) {
	ctx := context.TODO()
	data, err := d.readDataSource(ctx, alias, args...)
	if err != nil {
		return nil, err
	}
	return data.Unmarshal()
}

// DatasourceReachable - Determines if the named datasource is reachable with
// the given arguments. Reads from the datasource, and discards the returned data.
func (d *Data) DatasourceReachable(alias string, args ...string) bool {
	source := d.sourceReg.Get(alias)
	if source == nil {
		return false
	}
	ctx := context.TODO()
	_, err := source.Read(ctx, args...)
	return err == nil
}
