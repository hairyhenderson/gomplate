package main

import (
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
)

func init() {
	// Add some types we want to be able to handle which can be missing by default
	mime.AddExtensionType(".json", "application/json")
	mime.AddExtensionType(".yml", "application/yaml")
	mime.AddExtensionType(".yaml", "application/yaml")

	sourceReaders = make(map[string]func(*Source) ([]byte, error))

	// Register our source-reader functions
	addSourceReader("http", readHTTP)
	addSourceReader("https", readHTTP)
	addSourceReader("file", readFile)
}

var sourceReaders map[string]func(*Source) ([]byte, error)

// addSourceReader -
func addSourceReader(scheme string, readFunc func(*Source) ([]byte, error)) {
	sourceReaders[scheme] = readFunc
}

// Data -
type Data struct {
	Sources map[string]*Source
	cache   map[string][]byte
}

// NewData - constructor for Data
func NewData(datasourceArgs []string) *Data {
	sources := make(map[string]*Source)
	for _, v := range datasourceArgs {
		s, err := ParseSource(v)
		if err != nil {
			log.Fatalf("error parsing datasource %v", err)
			return nil
		}
		sources[s.Alias] = s
	}
	return &Data{
		Sources: sources,
	}
}

// Source - a data source
type Source struct {
	Alias string
	URL   *url.URL
	Ext   string
	Type  string
	FS    vfs.Filesystem // used for file: URLs, nil otherwise
	HC    *http.Client   // used for http[s]: URLs, nil otherwise
}

// NewSource - builds a &Source
func NewSource(alias string, URL *url.URL) (s *Source) {
	ext := filepath.Ext(URL.Path)

	var t string
	if ext != "" {
		t = mime.TypeByExtension(ext)
	}

	s = &Source{
		Alias: alias,
		URL:   URL,
		Ext:   ext,
		Type:  t,
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
		f := parts[0]
		alias = strings.SplitN(value, ".", 2)[0]
		if path.Base(f) != f {
			err := fmt.Errorf("Invalid datasource (%s). Must provide an alias with files not in working directory.", value)
			return nil, err
		}
		srcURL = absURL(f)
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
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Can't get working directory: %s", err)
	}
	baseURL := &url.URL{
		Scheme: "file",
		Path:   cwd + "/",
	}
	relURL := &url.URL{
		Path: value,
	}
	return baseURL.ResolveReference(relURL)
}

// Datasource -
func (d *Data) Datasource(alias string) map[string]interface{} {
	source := d.Sources[alias]
	b, err := d.ReadSource(source.FS, source)
	if err != nil {
		log.Fatalf("Couldn't read datasource '%s': %s", alias, err)
	}
	if source.Type == "application/json" {
		ty := &TypeConv{}
		return ty.JSON(string(b))
	}
	if source.Type == "application/yaml" {
		ty := &TypeConv{}
		return ty.YAML(string(b))
	}
	log.Fatalf("Datasources of type %s not yet supported", source.Type)
	return nil
}

// ReadSource -
func (d *Data) ReadSource(fs vfs.Filesystem, source *Source) ([]byte, error) {
	if d.cache == nil {
		d.cache = make(map[string][]byte)
	}
	cached, ok := d.cache[source.Alias]
	if ok {
		return cached, nil
	}
	if r, ok := sourceReaders[source.URL.Scheme]; ok {
		data, err := r(source)
		if err != nil {
			return nil, err
		}
		d.cache[source.Alias] = data
		return data, nil
	}

	log.Fatalf("Datasources with scheme %s not yet supported", source.URL.Scheme)
	return nil, nil
}

func readFile(source *Source) ([]byte, error) {
	if source.FS == nil {
		source.FS = vfs.OS()
	}

	// make sure we can access the file
	_, err := source.FS.Stat(source.URL.Path)
	if err != nil {
		log.Fatalf("Can't stat %s: %#v", source.URL.Path, err)
		return nil, err
	}

	f, err := source.FS.OpenFile(source.URL.Path, os.O_RDWR, 0)
	if err != nil {
		log.Fatalf("Can't open %s: %#v", source.URL.Path, err)
		return nil, err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("Can't read %s: %#v", source.URL.Path, err)
		return nil, err
	}
	return b, nil
}

func readHTTP(source *Source) ([]byte, error) {
	if source.HC == nil {
		source.HC = &http.Client{Timeout: time.Second * 5}
	}
	res, err := source.HC.Get(source.URL.String())
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		err := fmt.Errorf("Unexpected HTTP status %d on GET from %s: %s", res.StatusCode, source.URL, string(body))
		return nil, err
	}
	ctypeHdr := res.Header.Get("Content-Type")
	if ctypeHdr != "" {
		mediatype, _, e := mime.ParseMediaType(ctypeHdr)
		if e != nil {
			return nil, e
		}
		source.Type = mediatype
	}
	return body, nil
}
