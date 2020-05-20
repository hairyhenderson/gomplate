package datasource

import (
	"context"
	"fmt"
	"mime"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/hairyhenderson/gomplate/v3/internal/dataconv"
	"github.com/pkg/errors"
)

// Reader - Readers are used to read data from a data source, and can be
// re-used many times for different combinations of url/args.
type Reader interface {
	Read(ctx context.Context, url *url.URL, args ...string) (*Data, error)
}

// TODO: export Cleanup and move to Reader
type cleanerupper interface {
	cleanup()
}

// readers - map indexed by URL scheme
var readers map[string]Reader

func initReaders() {
	if readers == nil {
		readers = map[string]Reader{}
		readers["aws+smp"] = &AWSSMP{}
		readers["aws+sm"] = &AWSSecretsManager{}
		readers["boltdb"] = &BoltDB{}
		readers["env"] = &Env{}
		readers["file"] = &File{}
		readers["stdin"] = &Stdin{}
		readers["merge"] = &Merge{}

		consulReader := &Consul{}
		readers["consul"] = consulReader
		readers["consul+http"] = consulReader
		readers["consul+https"] = consulReader

		httpReader := &HTTP{}
		readers["http"] = httpReader
		readers["https"] = httpReader

		vaultReader := &Vault{}
		readers["vault"] = vaultReader
		readers["vault+http"] = vaultReader
		readers["vault+https"] = vaultReader

		blobReader := &Blob{}
		readers["s3"] = blobReader
		readers["gs"] = blobReader

		gitReader := &Git{}
		readers["git"] = gitReader
		readers["git+file"] = gitReader
		readers["git+http"] = gitReader
		readers["git+https"] = gitReader
		readers["git+ssh"] = gitReader
	}
}

// Data - some data read from a Reader.
type Data struct {
	Bytes   []byte
	URL     *url.URL
	Subpath string
	MType   string
}

func newData(url *url.URL, args []string) *Data {
	data := &Data{URL: url}
	if len(args) > 0 {
		data.Subpath = args[0]
	}
	return data
}

// MediaType returns the MIME type to use as a hint for parsing the datasource.
// It's expected that the datasource will have already been read before
// this function is called, and so the Data's mediaType property may be already set.
//
// The MIME type is determined by these rules:
// 1. the 'type' URL query parameter is used if present
// 2. otherwise, the mediaType property on the Data is used, if present
// 3. otherwise, a MIME type is calculated from the file extension, if the extension is registered
// 4. otherwise, the default type of 'text/plain' is used
func (d Data) MediaType() (mediatype string, err error) {
	subpath := d.Subpath
	if subpath != "" {
		if strings.HasPrefix(subpath, "//") {
			subpath = subpath[1:]
		}
		if !strings.HasPrefix(subpath, "/") {
			subpath = "/" + subpath
		}
	}
	subURL, err := url.Parse(subpath)
	if err != nil {
		return "", fmt.Errorf("couldn't parse subpath %q: %w", subpath, err)
	}
	mediatype = subURL.Query().Get("type")
	if mediatype == "" {
		mediatype = d.URL.Query().Get("type")
	}

	if mediatype == "" {
		mediatype = d.MType
	}

	// make it so + doesn't need to be escaped
	mediatype = strings.ReplaceAll(mediatype, " ", "+")

	if mediatype == "" {
		ext := filepath.Ext(subURL.Path)
		mediatype = mime.TypeByExtension(ext)
	}
	if mediatype == "" {
		ext := filepath.Ext(d.URL.Path)
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

// Unmarshal -
func (d *Data) Unmarshal() (out interface{}, err error) {
	mimeType, err := d.MediaType()
	if err != nil {
		return nil, err
	}

	s := string(d.Bytes)
	switch mimeAlias(mimeType) {
	case jsonMimetype:
		out, err = dataconv.JSON(s)
	case jsonArrayMimetype:
		out, err = dataconv.JSONArray(s)
	case yamlMimetype:
		out, err = dataconv.YAML(s)
	case csvMimetype:
		out, err = dataconv.CSV(s)
	case tomlMimetype:
		out, err = dataconv.TOML(s)
	case envMimetype:
		out, err = dataconv.DotEnv(s)
	case textMimetype:
		out = s
	default:
		return nil, errors.Errorf("Datasources of type %s not yet supported", mimeType)
	}
	return out, err
}
