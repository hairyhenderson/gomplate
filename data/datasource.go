package data

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"

	"github.com/pkg/errors"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/hairyhenderson/gomplate/v3/libkv"
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

// registerReaders registers the source-reader functions
func (d *Data) registerReaders() {
	d.sourceReaders = make(map[string]func(context.Context, *Source, ...string) ([]byte, error))

	d.sourceReaders["aws+smp"] = readAWSSMP
	d.sourceReaders["aws+sm"] = readAWSSecretsManager
	d.sourceReaders["consul"] = readConsul
	d.sourceReaders["consul+http"] = readConsul
	d.sourceReaders["consul+https"] = readConsul
	d.sourceReaders["env"] = readEnv
	d.sourceReaders["file"] = readFile
	d.sourceReaders["http"] = readHTTP
	d.sourceReaders["https"] = readHTTP
	d.sourceReaders["merge"] = d.readMerge
	d.sourceReaders["stdin"] = readStdin
	d.sourceReaders["vault"] = readVault
	d.sourceReaders["vault+http"] = readVault
	d.sourceReaders["vault+https"] = readVault
	d.sourceReaders["s3"] = readBlob
	d.sourceReaders["gs"] = readBlob
	d.sourceReaders["git"] = readGit
	d.sourceReaders["git+file"] = readGit
	d.sourceReaders["git+http"] = readGit
	d.sourceReaders["git+https"] = readGit
	d.sourceReaders["git+ssh"] = readGit
}

// lookupReader - return the reader function for the given scheme
func (d *Data) lookupReader(scheme string) (func(context.Context, *Source, ...string) ([]byte, error), error) {
	if d.sourceReaders == nil {
		d.registerReaders()
	}
	r, ok := d.sourceReaders[scheme]
	if !ok {
		return nil, errors.Errorf("scheme %s not registered", scheme)
	}
	return r, nil
}

// Data -
// Deprecated: will be replaced in future
type Data struct {
	Ctx context.Context

	Sources map[string]*Source

	sourceReaders map[string]func(context.Context, *Source, ...string) ([]byte, error)
	cache         map[string][]byte

	// headers from the --datasource-header/-H option that don't reference datasources from the commandline
	ExtraHeaders map[string]http.Header
}

// Cleanup - clean up datasources before shutting the process down - things
// like Logging out happen here
func (d *Data) Cleanup() {
	for _, s := range d.Sources {
		s.cleanup()
	}
}

// NewData - constructor for Data
// Deprecated: will be replaced in future
func NewData(datasourceArgs, headerArgs []string) (*Data, error) {
	cfg := &config.Config{}
	err := cfg.ParseDataSourceFlags(datasourceArgs, nil, nil, headerArgs)
	if err != nil {
		return nil, err
	}
	data := FromConfig(context.Background(), cfg)
	return data, nil
}

// FromConfig - internal use only!
func FromConfig(ctx context.Context, cfg *config.Config) *Data {
	// XXX: This is temporary, and will be replaced with something a bit cleaner
	// when datasources are refactored
	ctx = ContextWithStdin(ctx, cfg.Stdin)

	sources := map[string]*Source{}
	for alias, d := range cfg.DataSources {
		sources[alias] = &Source{
			Alias:  alias,
			URL:    d.URL,
			Header: d.Header,
		}
	}
	for alias, d := range cfg.Context {
		sources[alias] = &Source{
			Alias:  alias,
			URL:    d.URL,
			Header: d.Header,
		}
	}
	return &Data{
		Ctx:          ctx,
		Sources:      sources,
		ExtraHeaders: cfg.ExtraHeaders,
	}
}

// Source - a data source
// Deprecated: will be replaced in future
type Source struct {
	Alias             string
	URL               *url.URL
	Header            http.Header             // used for http[s]: URLs, nil otherwise
	fs                afero.Fs                // used for file: URLs, nil otherwise
	hc                *http.Client            // used for http[s]: URLs, nil otherwise
	vc                *vault.Vault            // used for vault: URLs, nil otherwise
	kv                *libkv.LibKV            // used for consul:, etcd:, zookeeper: URLs, nil otherwise
	asmpg             awssmpGetter            // used for aws+smp:, nil otherwise
	awsSecretsManager awsSecretsManagerGetter // used for aws+sm, nil otherwise
	mediaType         string
}

func (s *Source) inherit(parent *Source) {
	s.fs = parent.fs
	s.hc = parent.hc
	s.vc = parent.vc
	s.kv = parent.kv
	s.asmpg = parent.asmpg
}

func (s *Source) cleanup() {
	if s.vc != nil {
		s.vc.Logout()
	}
	if s.kv != nil {
		s.kv.Logout()
	}
}

// mimeType returns the MIME type to use as a hint for parsing the datasource.
// It's expected that the datasource will have already been read before
// this function is called, and so the Source's Type property may be already set.
//
// The MIME type is determined by these rules:
// 1. the 'type' URL query parameter is used if present
// 2. otherwise, the Type property on the Source is used, if present
// 3. otherwise, a MIME type is calculated from the file extension, if the extension is registered
// 4. otherwise, the default type of 'text/plain' is used
func (s *Source) mimeType(arg string) (mimeType string, err error) {
	if len(arg) > 0 {
		if strings.HasPrefix(arg, "//") {
			arg = arg[1:]
		}
		if !strings.HasPrefix(arg, "/") {
			arg = "/" + arg
		}
	}
	argURL, err := url.Parse(arg)
	if err != nil {
		return "", fmt.Errorf("mimeType: couldn't parse arg %q: %w", arg, err)
	}
	mediatype := argURL.Query().Get("type")
	if mediatype == "" {
		mediatype = s.URL.Query().Get("type")
	}

	if mediatype == "" {
		mediatype = s.mediaType
	}

	// make it so + doesn't need to be escaped
	mediatype = strings.ReplaceAll(mediatype, " ", "+")

	if mediatype == "" {
		ext := filepath.Ext(argURL.Path)
		mediatype = mime.TypeByExtension(ext)
	}

	if mediatype == "" {
		ext := filepath.Ext(s.URL.Path)
		mediatype = mime.TypeByExtension(ext)
	}

	if mediatype != "" {
		t, _, err := mime.ParseMediaType(mediatype)
		if err != nil {
			return "", errors.Wrapf(err, "MIME type was %q", mediatype)
		}
		mediatype = t
		return mediatype, nil
	}

	return textMimetype, nil
}

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (s *Source) String() string {
	return fmt.Sprintf("%s=%s (%s)", s.Alias, s.URL.String(), s.mediaType)
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
	s := &Source{
		Alias:  alias,
		URL:    srcURL,
		Header: d.ExtraHeaders[alias],
	}
	if d.Sources == nil {
		d.Sources = make(map[string]*Source)
	}
	d.Sources[alias] = s
	return "", nil
}

// DatasourceExists -
func (d *Data) DatasourceExists(alias string) bool {
	_, ok := d.Sources[alias]
	return ok
}

func (d *Data) lookupSource(alias string) (*Source, error) {
	source, ok := d.Sources[alias]
	if !ok {
		srcURL, err := url.Parse(alias)
		if err != nil || !srcURL.IsAbs() {
			return nil, errors.Errorf("Undefined datasource '%s'", alias)
		}
		source = &Source{
			Alias:  alias,
			URL:    srcURL,
			Header: d.ExtraHeaders[alias],
		}
		d.Sources[alias] = source
	}
	if source.Alias == "" {
		source.Alias = alias
	}
	return source, nil
}

func (d *Data) readDataSource(ctx context.Context, alias string, args ...string) (data, mimeType string, err error) {
	source, err := d.lookupSource(alias)
	if err != nil {
		return "", "", err
	}
	b, err := d.readSource(ctx, source, args...)
	if err != nil {
		return "", "", errors.Wrapf(err, "Couldn't read datasource '%s'", alias)
	}

	subpath := ""
	if len(args) > 0 {
		subpath = args[0]
	}
	mimeType, err = source.mimeType(subpath)
	if err != nil {
		return "", "", err
	}
	return string(b), mimeType, nil
}

// Include -
func (d *Data) Include(alias string, args ...string) (string, error) {
	data, _, err := d.readDataSource(d.Ctx, alias, args...)
	return data, err
}

// Datasource -
func (d *Data) Datasource(alias string, args ...string) (interface{}, error) {
	data, mimeType, err := d.readDataSource(d.Ctx, alias, args...)
	if err != nil {
		return nil, err
	}

	return parseData(mimeType, data)
}

func parseData(mimeType, s string) (out interface{}, err error) {
	switch mimeAlias(mimeType) {
	case jsonMimetype:
		out, err = JSON(s)
		if err != nil {
			// maybe it's a JSON array
			out, err = JSONArray(s)
		}
	case jsonArrayMimetype:
		out, err = JSONArray(s)
	case yamlMimetype:
		out, err = YAML(s)
		if err != nil {
			// maybe it's a YAML array
			out, err = YAMLArray(s)
		}
	case csvMimetype:
		out, err = CSV(s)
	case tomlMimetype:
		out, err = TOML(s)
	case envMimetype:
		out, err = dotEnv(s)
	case textMimetype:
		out = s
	default:
		return nil, errors.Errorf("Datasources of type %s not yet supported", mimeType)
	}
	return out, err
}

// DatasourceReachable - Determines if the named datasource is reachable with
// the given arguments. Reads from the datasource, and discards the returned data.
func (d *Data) DatasourceReachable(alias string, args ...string) bool {
	source, ok := d.Sources[alias]
	if !ok {
		return false
	}
	_, err := d.readSource(d.Ctx, source, args...)
	return err == nil
}

// readSource returns the (possibly cached) data from the given source,
// as referenced by the given args
func (d *Data) readSource(ctx context.Context, source *Source, args ...string) ([]byte, error) {
	if d.cache == nil {
		d.cache = make(map[string][]byte)
	}
	cacheKey := source.Alias
	for _, v := range args {
		cacheKey += v
	}
	cached, ok := d.cache[cacheKey]
	if ok {
		return cached, nil
	}
	r, err := d.lookupReader(source.URL.Scheme)
	if err != nil {
		return nil, errors.Wrap(err, "Datasource not yet supported")
	}
	data, err := r(ctx, source, args...)
	if err != nil {
		return nil, err
	}
	d.cache[cacheKey] = data
	return data, nil
}

// Show all datasources  -
func (d *Data) ListDatasources() []string {
	datasources := make([]string, 0, len(d.Sources))
	for source := range d.Sources {
		datasources = append(datasources, source)
	}
	sort.Strings(datasources)
	return datasources
}
