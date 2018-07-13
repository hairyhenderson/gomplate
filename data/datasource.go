package data

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/service/ssm"

	"github.com/blang/vfs"
	"github.com/hairyhenderson/gomplate/libkv"
	"github.com/hairyhenderson/gomplate/vault"
)

// stdin - for overriding in tests
var stdin io.Reader

func regExtension(ext, typ string) {
	err := mime.AddExtensionType(ext, typ)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	// Add some types we want to be able to handle which can be missing by default
	regExtension(".json", jsonMimetype)
	regExtension(".yml", yamlMimetype)
	regExtension(".yaml", yamlMimetype)
	regExtension(".csv", csvMimetype)
	regExtension(".toml", tomlMimetype)

	sourceReaders = make(map[string]func(*Source, ...string) ([]byte, error))

	// Register our source-reader functions
	addSourceReader("http", readHTTP)
	addSourceReader("https", readHTTP)
	addSourceReader("file", readFile)
	addSourceReader("stdin", readStdin)
	addSourceReader("vault", readVault)
	addSourceReader("vault+http", readVault)
	addSourceReader("vault+https", readVault)
	addSourceReader("consul", readConsul)
	addSourceReader("consul+http", readConsul)
	addSourceReader("consul+https", readConsul)
	addSourceReader("boltdb", readBoltDB)
	addSourceReader("aws+smp", readAWSSMP)
}

var sourceReaders map[string]func(*Source, ...string) ([]byte, error)

// addSourceReader -
func addSourceReader(scheme string, readFunc func(*Source, ...string) ([]byte, error)) {
	sourceReaders[scheme] = readFunc
}

// Data -
type Data struct {
	Sources map[string]*Source
	cache   map[string][]byte

	// headers from the --datasource-header/-H option that don't reference datasources from the commandline
	extraHeaders map[string]http.Header
}

// Cleanup - clean up datasources before shutting the process down - things
// like Logging out happen here
func (d *Data) Cleanup() {
	for _, s := range d.Sources {
		s.cleanup()
	}
}

// NewData - constructor for Data
func NewData(datasourceArgs []string, headerArgs []string) (*Data, error) {
	sources := make(map[string]*Source)
	headers, err := parseHeaderArgs(headerArgs)
	if err != nil {
		return nil, err
	}
	for _, v := range datasourceArgs {
		s, err := ParseSource(v)
		if err != nil {
			return nil, errors.Wrapf(err, "error parsing datasource")
		}
		s.Header = headers[s.Alias]
		// pop the header out of the map, so we end up with only the unreferenced ones
		delete(headers, s.Alias)

		sources[s.Alias] = s
	}
	return &Data{
		Sources:      sources,
		extraHeaders: headers,
	}, nil
}

// AWSSMPGetter - A subset of SSM API for use in unit testing
type AWSSMPGetter interface {
	GetParameter(*ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
}

// Source - a data source
type Source struct {
	Alias  string
	URL    *url.URL
	Ext    string
	Type   string
	Params map[string]string
	FS     vfs.Filesystem // used for file: URLs, nil otherwise
	HC     *http.Client   // used for http[s]: URLs, nil otherwise
	VC     *vault.Vault   // used for vault: URLs, nil otherwise
	KV     *libkv.LibKV   // used for consul:, etcd:, zookeeper: & boltdb: URLs, nil otherwise
	ASMPG  AWSSMPGetter   // used for aws+smp:, nil otherwise
	Header http.Header    // used for http[s]: URLs, nil otherwise
}

func (s *Source) cleanup() {
	if s.VC != nil {
		s.VC.Logout()
	}
	if s.KV != nil {
		s.KV.Logout()
	}
}

// NewSource - builds a &Source
func NewSource(alias string, URL *url.URL) (*Source, error) {
	ext := filepath.Ext(URL.Path)

	s := &Source{
		Alias: alias,
		URL:   URL,
		Ext:   ext,
	}

	mediatype := s.URL.Query().Get("type")
	if mediatype == "" {
		mediatype = mime.TypeByExtension(ext)
	}
	if mediatype != "" {
		t, params, err := mime.ParseMediaType(mediatype)
		if err != nil {
			return nil, err
		}
		s.Type = t
		s.Params = params
	}
	if s.Type == "" {
		s.Type = textMimetype
	}
	return s, nil
}

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (s *Source) String() string {
	return fmt.Sprintf("%s=%s (%s)", s.Alias, s.URL.String(), s.Type)
}

// ParseSource -
func ParseSource(value string) (*Source, error) {
	var (
		alias  string
		srcURL *url.URL
		err    error
	)
	parts := strings.SplitN(value, "=", 2)
	if len(parts) == 1 {
		f := parts[0]
		alias = strings.SplitN(value, ".", 2)[0]
		if path.Base(f) != f {
			err = errors.Errorf("Invalid datasource (%s). Must provide an alias with files not in working directory", value)
			return nil, err
		}
		srcURL, err = absURL(f)
		if err != nil {
			return nil, err
		}
	} else if len(parts) == 2 {
		alias = parts[0]
		srcURL, err = parseSourceURL(parts[1])
		if err != nil {
			return nil, err
		}
	}

	return NewSource(alias, srcURL)
}

func parseSourceURL(value string) (*url.URL, error) {
	if value == "-" {
		value = "stdin://"
	}
	srcURL, err := url.Parse(value)
	if err != nil {
		return nil, err
	}

	if !srcURL.IsAbs() {
		srcURL, err = absURL(value)
		if err != nil {
			return nil, err
		}
	}
	return srcURL, nil
}

func absURL(value string) (*url.URL, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrapf(err, "can't get working directory")
	}
	urlCwd := strings.Replace(cwd, string(os.PathSeparator), "/", -1)
	baseURL := &url.URL{
		Scheme: "file",
		Path:   urlCwd + "/",
	}
	relURL := &url.URL{
		Path: value,
	}
	return baseURL.ResolveReference(relURL), nil
}

// DefineDatasource -
func (d *Data) DefineDatasource(alias, value string) (*Source, error) {
	if alias == "" {
		return nil, errors.New("datasource alias must be provided")
	}
	if d.DatasourceExists(alias) {
		return d.Sources[alias], nil
	}
	srcURL, err := parseSourceURL(value)
	if err != nil {
		return nil, err
	}
	s, err := NewSource(alias, srcURL)
	if err != nil {
		return nil, err
	}
	s.Header = d.extraHeaders[s.Alias]
	if d.Sources == nil {
		d.Sources = make(map[string]*Source)
	}
	d.Sources[s.Alias] = s
	return s, nil
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
		source, err = NewSource(alias, srcURL)
		if err != nil {
			return nil, err
		}
		source.Header = d.extraHeaders[alias]
		d.Sources[alias] = source
	}
	return source, nil
}

// Datasource -
func (d *Data) Datasource(alias string, args ...string) (interface{}, error) {
	source, err := d.lookupSource(alias)
	if err != nil {
		return nil, err
	}
	b, err := d.ReadSource(source, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "Couldn't read datasource '%s'", alias)
	}
	s := string(b)
	var out interface{}
	switch source.Type {
	case jsonMimetype:
		out = JSON(s)
	case jsonArrayMimetype:
		out = JSONArray(s)
	case yamlMimetype:
		out = YAML(s)
	case csvMimetype:
		out = CSV(s)
	case tomlMimetype:
		out = TOML(s)
	case textMimetype:
		out = s
	default:
		return nil, errors.Errorf("Datasources of type %s not yet supported", source.Type)
	}
	return out, nil
}

// DatasourceReachable - Determines if the named datasource is reachable with
// the given arguments. Reads from the datasource, and discards the returned data.
func (d *Data) DatasourceReachable(alias string, args ...string) bool {
	source, ok := d.Sources[alias]
	if !ok {
		return false
	}
	_, err := d.ReadSource(source, args...)
	return err == nil
}

// Include -
func (d *Data) Include(alias string, args ...string) (string, error) {
	source, ok := d.Sources[alias]
	if !ok {
		return "", errors.Errorf("Undefined datasource '%s'", alias)
	}
	b, err := d.ReadSource(source, args...)
	if err != nil {
		return "", errors.Wrapf(err, "Couldn't read datasource '%s'", alias)
	}
	return string(b), nil
}

// ReadSource -
func (d *Data) ReadSource(source *Source, args ...string) ([]byte, error) {
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
	if r, ok := sourceReaders[source.URL.Scheme]; ok {
		data, err := r(source, args...)
		if err != nil {
			return nil, err
		}
		d.cache[cacheKey] = data
		return data, nil
	}

	return nil, errors.Errorf("Datasources with scheme %s not yet supported", source.URL.Scheme)
}

func readStdin(source *Source, args ...string) ([]byte, error) {
	if stdin == nil {
		stdin = os.Stdin
	}
	b, err := ioutil.ReadAll(stdin)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't read %s", stdin)
	}
	return b, nil
}

func readHTTP(source *Source, args ...string) ([]byte, error) {
	if source.HC == nil {
		source.HC = &http.Client{Timeout: time.Second * 5}
	}
	req, err := http.NewRequest("GET", source.URL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header = source.Header
	res, err := source.HC.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		err := errors.Errorf("Unexpected HTTP status %d on GET from %s: %s", res.StatusCode, source.URL, string(body))
		return nil, err
	}
	ctypeHdr := res.Header.Get("Content-Type")
	if ctypeHdr != "" {
		mediatype, params, e := mime.ParseMediaType(ctypeHdr)
		if e != nil {
			return nil, e
		}
		source.Type = mediatype
		source.Params = params
	}
	return body, nil
}

func readConsul(source *Source, args ...string) ([]byte, error) {
	if source.KV == nil {
		source.KV = libkv.NewConsul(source.URL)
		err := source.KV.Login()
		if err != nil {
			return nil, err
		}
	}

	p := source.URL.Path
	if len(args) == 1 {
		p = p + "/" + args[0]
	}

	data, err := source.KV.Read(p)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func readBoltDB(source *Source, args ...string) ([]byte, error) {
	if source.KV == nil {
		source.KV = libkv.NewBoltDB(source.URL)
	}

	if len(args) != 1 {
		return nil, errors.New("missing key")
	}
	p := args[0]

	data, err := source.KV.Read(p)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func parseHeaderArgs(headerArgs []string) (map[string]http.Header, error) {
	headers := make(map[string]http.Header)
	for _, v := range headerArgs {
		ds, name, value, err := splitHeaderArg(v)
		if err != nil {
			return nil, err
		}
		if _, ok := headers[ds]; !ok {
			headers[ds] = make(http.Header)
		}
		headers[ds][name] = append(headers[ds][name], strings.TrimSpace(value))
	}
	return headers, nil
}

func splitHeaderArg(arg string) (datasourceAlias, name, value string, err error) {
	parts := strings.SplitN(arg, "=", 2)
	if len(parts) != 2 {
		err = errors.Errorf("Invalid datasource-header option '%s'", arg)
		return "", "", "", err
	}
	datasourceAlias = parts[0]
	name, value, err = splitHeader(parts[1])
	return datasourceAlias, name, value, err
}

func splitHeader(header string) (name, value string, err error) {
	parts := strings.SplitN(header, ":", 2)
	if len(parts) != 2 {
		err = errors.Errorf("Invalid HTTP Header format '%s'", header)
		return "", "", err
	}
	name = http.CanonicalHeaderKey(parts[0])
	value = parts[1]
	return name, value, nil
}
