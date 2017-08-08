package main

import (
	"errors"
	"fmt"
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

	"github.com/blang/vfs"
	"github.com/hairyhenderson/gomplate/libkv"
	"github.com/hairyhenderson/gomplate/vault"
)

// logFatal is defined so log.Fatal calls can be overridden for testing
var logFatalf = log.Fatalf

func regExtension(ext, typ string) {
	err := mime.AddExtensionType(ext, typ)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	// Add some types we want to be able to handle which can be missing by default
	regExtension(".json", "application/json")
	regExtension(".yml", "application/yaml")
	regExtension(".yaml", "application/yaml")
	regExtension(".csv", "text/csv")
	regExtension(".toml", "application/toml")

	sourceReaders = make(map[string]func(*Source, ...string) ([]byte, error))

	// Register our source-reader functions
	addSourceReader("http", readHTTP)
	addSourceReader("https", readHTTP)
	addSourceReader("file", readFile)
	addSourceReader("vault", readVault)
	addSourceReader("consul", readConsul)
	addSourceReader("consul+http", readConsul)
	addSourceReader("consul+https", readConsul)
	addSourceReader("boltdb", readBoltDB)
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
}

// NewData - constructor for Data
func NewData(datasourceArgs []string, headerArgs []string) *Data {
	sources := make(map[string]*Source)
	headers := parseHeaderArgs(headerArgs)
	for _, v := range datasourceArgs {
		for _, arg := range fanOutDatasourceArgs(v) {
			s, err := ParseSource(arg)
			if err != nil {
				log.Fatalf("error parsing datasource %s: %v", arg, err)
				return nil
			}
			s.Header = headers[s.Alias]
			sources[s.Alias] = s
		}
	}
	return &Data{
		Sources: sources,
	}
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
	Header http.Header    // used for http[s]: URLs, nil otherwise
}

// NewSource - builds a &Source
func NewSource(alias string, URL *url.URL) (s *Source) {
	ext := filepath.Ext(URL.Path)

	s = &Source{
		Alias: alias,
		URL:   URL,
		Ext:   ext,
	}

	if ext != "" && URL.Scheme != "boltdb" {
		mediatype := mime.TypeByExtension(ext)
		t, params, err := mime.ParseMediaType(mediatype)
		if err != nil {
			log.Fatal(err)
		}
		s.Type = t
		s.Params = params
	}
	return
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
	)
	parts := strings.SplitN(value, "=", 2)
	if len(parts) == 1 {
		srcURL = absURL(value)
		if srcURL.Scheme == "file" {
			alias = strings.SplitN(path.Base(value), ".", 2)[0]
		} else {
			err := fmt.Errorf("Invalid datasource (%s). Must provide an alias name for non-file datasources", value)
			return nil, err
		}
	} else if len(parts) == 2 {
		alias = parts[0]
		var err error
		srcURL, err = url.Parse(parts[1])
		if err != nil {
			return nil, err
		}

		if !srcURL.IsAbs() {
			srcURL = absURL(parts[1])
		}
	}

	s := NewSource(alias, srcURL)
	return s, nil
}

func absURL(value string) *url.URL {
	parsedURL, err := url.Parse(value)
	if err != nil || parsedURL.Scheme == "" || !parsedURL.IsAbs() {
		return fileToURL(value)
	}
	return parsedURL
}

func fileToURL(path string) *url.URL {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Can't get working directory: %s", err)
	}
	urlCwd := strings.Replace(cwd, string(os.PathSeparator), "/", -1)
	baseURL := &url.URL{
		Scheme: "file",
		Path:   urlCwd + "/",
	}
	relURL := &url.URL{
		Path: path,
	}
	return baseURL.ResolveReference(relURL)
}

// DatasourceExists -
func (d *Data) DatasourceExists(alias string) bool {
	_, ok := d.Sources[alias]
	return ok
}

const plaintext = "text/plain"

// Datasource -
func (d *Data) Datasource(alias string, args ...string) interface{} {
	source, ok := d.Sources[alias]
	if !ok {
		log.Fatalf("Undefined datasource '%s'", alias)
	}
	b, err := d.ReadSource(source, args...)
	if err != nil {
		log.Fatalf("Couldn't read datasource '%s': %s", alias, err)
	}
	s := string(b)
	ty := &TypeConv{}
	if source.Type == "application/json" {
		return ty.JSON(s)
	}
	if source.Type == "application/yaml" {
		return ty.YAML(s)
	}
	if source.Type == "text/csv" {
		return ty.CSV(s)
	}
	if source.Type == "application/toml" {
		return ty.TOML(s)
	}
	if source.Type == plaintext {
		return s
	}
	log.Fatalf("Datasources of type %s not yet supported", source.Type)
	return nil
}

// Include -
func (d *Data) include(alias string, args ...string) interface{} {
	source, ok := d.Sources[alias]
	if !ok {
		log.Fatalf("Undefined datasource '%s'", alias)
	}
	b, err := d.ReadSource(source, args...)
	if err != nil {
		log.Fatalf("Couldn't read datasource '%s': %s", alias, err)
	}
	return string(b)
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

	log.Fatalf("Datasources with scheme %s not yet supported", source.URL.Scheme)
	return nil, nil
}

// If argument is a plain directory, use dir contents as datasource files
func fanOutDatasourceArgs(arg string) []string {
	if _, err := os.Stat(arg); err == nil {
		files, err := ioutil.ReadDir(arg)
		if err != nil {
			log.Fatalf("error parsing datasource directory %s %v", arg, err)
			return nil
		}

		var paths []string
		for _, file := range files {
			if !file.IsDir() {
				paths = append(paths, path.Join(arg, file.Name()))
			}
		}
		return paths
	}
	return []string{arg}
}

func readFile(source *Source, args ...string) ([]byte, error) {
	if source.FS == nil {
		source.FS = vfs.OS()
	}

	p := filepath.FromSlash(source.URL.Path)

	// make sure we can access the file
	_, err := source.FS.Stat(p)
	if err != nil {
		log.Fatalf("Can't stat %s: %#v", p, err)
		return nil, err
	}

	f, err := source.FS.OpenFile(p, os.O_RDONLY, 0)
	if err != nil {
		log.Fatalf("Can't open %s: %#v", p, err)
		return nil, err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("Can't read %s: %#v", p, err)
		return nil, err
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
		err := fmt.Errorf("Unexpected HTTP status %d on GET from %s: %s", res.StatusCode, source.URL, string(body))
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

func readVault(source *Source, args ...string) ([]byte, error) {
	if source.VC == nil {
		source.VC = vault.New()
		source.VC.Login()
		addCleanupHook(source.VC.Logout)
	}

	params := make(map[string]interface{})

	p := source.URL.Path

	for key, val := range source.URL.Query() {
		params[key] = strings.Join(val, " ")
	}

	if len(args) == 1 {
		parsed, err := url.Parse(args[0])
		if err != nil {
			return nil, err
		}

		if parsed.Path != "" {
			p = p + "/" + parsed.Path
		}

		for key, val := range parsed.Query() {
			params[key] = strings.Join(val, " ")
		}
	}

	var data []byte
	var err error

	if len(params) > 0 {
		data, err = source.VC.Write(p, params)
	} else {
		data, err = source.VC.Read(p)
	}
	if err != nil {
		return nil, err
	}
	source.Type = "application/json"

	return data, nil
}

func readConsul(source *Source, args ...string) ([]byte, error) {
	if source.KV == nil {
		source.KV = libkv.NewConsul(source.URL)
		err := source.KV.Login()
		addCleanupHook(source.KV.Logout)
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
	source.Type = plaintext

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
	source.Type = plaintext

	return data, nil
}

func parseHeaderArgs(headerArgs []string) map[string]http.Header {
	headers := make(map[string]http.Header)
	for _, v := range headerArgs {
		ds, name, value := splitHeaderArg(v)
		if _, ok := headers[ds]; !ok {
			headers[ds] = make(http.Header)
		}
		headers[ds][name] = append(headers[ds][name], strings.TrimSpace(value))
	}
	return headers
}

func splitHeaderArg(arg string) (datasourceAlias, name, value string) {
	parts := strings.SplitN(arg, "=", 2)
	if len(parts) != 2 {
		logFatalf("Invalid datasource-header option '%s'", arg)
		return "", "", ""
	}
	datasourceAlias = parts[0]
	name, value = splitHeader(parts[1])
	return datasourceAlias, name, value
}

func splitHeader(header string) (name, value string) {
	parts := strings.SplitN(header, ":", 2)
	if len(parts) != 2 {
		logFatalf("Invalid HTTP Header format '%s'", header)
		return "", ""
	}
	name = http.CanonicalHeaderKey(parts[0])
	value = parts[1]
	return name, value
}
