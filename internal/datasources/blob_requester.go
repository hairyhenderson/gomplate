package datasources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	gaws "github.com/hairyhenderson/gomplate/v3/aws"
	"github.com/hairyhenderson/gomplate/v3/env"
	"github.com/pkg/errors"

	"gocloud.dev/blob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/s3blob"
	"gocloud.dev/gcp"
)

const (
	schemeS3 = "s3"
	schemeGS = "gs"
)

type blobRequester struct{}

func (r *blobRequester) Request(ctx context.Context, u *url.URL, header http.Header) (*Response, error) {
	key := u.Path

	opener, err := r.newOpener(ctx, u.Scheme)
	if err != nil {
		return nil, err
	}

	mux := blob.URLMux{}
	mux.RegisterBucket(u.Scheme, opener)

	bucket, err := mux.OpenBucket(ctx, sanitizeURL(u).String())
	if err != nil {
		return nil, err
	}
	defer bucket.Close()

	resp := &Response{}
	if strings.HasSuffix(key, "/") {
		b, lerr := r.listBucket(ctx, bucket, key)
		if lerr != nil {
			return nil, lerr
		}
		resp.ContentType = jsonArrayMimetype
		resp.ContentLength = int64(b.Len())
		resp.Body = ioutil.NopCloser(b)
	} else {
		attr, data, gerr := r.getBlob(ctx, bucket, key)
		if gerr != nil {
			return nil, gerr
		}
		resp.ContentType = attr.ContentType
		resp.ContentLength = attr.Size
		resp.Body = data
	}

	// try to guess the right media, but fall back to resp.ContentType if it
	// can't be guessed
	if mt, merr := mimeType(u, resp.ContentType); merr == nil {
		resp.ContentType = mt
	}

	return resp, err
}

// create the correct kind of blob.BucketURLOpener for the given URL
func (r *blobRequester) newOpener(ctx context.Context, scheme string) (opener blob.BucketURLOpener, err error) {
	switch scheme {
	case schemeS3:
		// set up a "regular" gomplate AWS SDK session
		sess := gaws.SDKSession()
		// see https://gocloud.dev/concepts/urls/#muxes
		opener = &s3blob.URLOpener{ConfigProvider: sess}
	case schemeGS:
		creds, err := gcp.DefaultCredentials(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve GCP credentials")
		}

		client, err := gcp.NewHTTPClient(
			gcp.DefaultTransport(),
			gcp.CredentialsTokenSource(creds))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create GCP HTTP client")
		}
		opener = &gcsblob.URLOpener{
			Client: client,
		}
	}
	return opener, nil
}

func (r *blobRequester) getBlob(ctx context.Context, bucket *blob.Bucket, key string) (attr *blob.Attributes, data *blob.Reader, err error) {
	key = strings.TrimPrefix(key, "/")
	attr, err = bucket.Attributes(ctx, key)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to retrieve attributes for %s", key)
	}
	rdr, err := bucket.NewReader(ctx, key, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open blob reader for %s: %w", key, err)
	}
	return attr, rdr, nil
}

// calls the bucket listing API, returning a JSON Array
func (r *blobRequester) listBucket(ctx context.Context, bucket *blob.Bucket, path string) (*bytes.Buffer, error) {
	path = strings.TrimPrefix(path, "/")
	opts := &blob.ListOptions{
		Prefix:    path,
		Delimiter: "/",
	}
	li := bucket.List(opts)
	keys := []string{}
	for {
		obj, err := li.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		keys = append(keys, strings.TrimPrefix(obj.Key, path))
	}

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(keys); err != nil {
		return nil, err
	}
	b := buf.Bytes()
	// chop off the newline added by the json encoder
	data := bytes.NewBuffer(b[:len(b)-1])
	return data, nil
}

// copy/sanitize the URL for the Go CDK - it doesn't like params it can't parse
func sanitizeURL(u *url.URL) *url.URL {
	out := cloneURL(u)
	q := out.Query()

	for param := range q {
		switch u.Scheme {
		case schemeS3:
			switch param {
			case "region", "endpoint", "disableSSL", "s3ForcePathStyle":
			default:
				q.Del(param)
			}
		case schemeGS:
			switch param {
			case "access_id", "private_key_path":
			default:
				q.Del(param)
			}
		}
	}

	if u.Scheme == schemeS3 {
		// handle AWS_S3_ENDPOINT env var
		endpoint := env.Getenv("AWS_S3_ENDPOINT")
		if endpoint != "" {
			q.Set("endpoint", endpoint)
		}
	}

	out.RawQuery = q.Encode()

	return out
}
