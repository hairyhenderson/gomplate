package data

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/url"
	"path"
	"strings"

	gaws "github.com/flanksource/gomplate/v3/aws"
	"github.com/flanksource/gomplate/v3/env"
	"github.com/pkg/errors"

	"gocloud.dev/blob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/s3blob"
	"gocloud.dev/gcp"
)

func readBlob(ctx context.Context, source *Source, args ...string) (output []byte, err error) {
	if len(args) >= 2 {
		return nil, errors.New("maximum two arguments to blob datasource: alias, extraPath")
	}

	key := source.URL.Path
	if len(args) == 1 {
		key = path.Join(key, args[0])
	}

	opener, err := newOpener(ctx, source.URL)
	if err != nil {
		return nil, err
	}

	mux := blob.URLMux{}
	mux.RegisterBucket(source.URL.Scheme, opener)

	u := blobURL(source.URL)
	bucket, err := mux.OpenBucket(ctx, u)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()

	var r func(context.Context, *blob.Bucket, string) (string, []byte, error)
	if strings.HasSuffix(key, "/") {
		r = listBucket
	} else {
		r = getBlob
	}

	mediaType, data, err := r(ctx, bucket, key)
	if mediaType != "" {
		source.mediaType = mediaType
	}
	return data, err
}

// create the correct kind of blob.BucketURLOpener for the given URL
func newOpener(ctx context.Context, u *url.URL) (opener blob.BucketURLOpener, err error) {
	switch u.Scheme {
	case "s3":
		// set up a "regular" gomplate AWS SDK session
		sess := gaws.SDKSession()
		// see https://gocloud.dev/concepts/urls/#muxes
		opener = &s3blob.URLOpener{ConfigProvider: sess}
	case "gs":
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

func getBlob(ctx context.Context, bucket *blob.Bucket, key string) (mediaType string, data []byte, err error) {
	key = strings.TrimPrefix(key, "/")
	attr, err := bucket.Attributes(ctx, key)
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to retrieve attributes for %s", key)
	}
	if attr.ContentType != "" {
		mt, _, e := mime.ParseMediaType(attr.ContentType)
		if e != nil {
			return "", nil, e
		}
		mediaType = mt
	}
	data, err = bucket.ReadAll(ctx, key)
	return mediaType, data, errors.Wrapf(err, "failed to read %s", key)
}

// calls the bucket listing API, returning a JSON Array
func listBucket(ctx context.Context, bucket *blob.Bucket, path string) (mediaType string, data []byte, err error) {
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
			return "", nil, err
		}
		keys = append(keys, strings.TrimPrefix(obj.Key, path))
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(keys); err != nil {
		return "", nil, err
	}
	b := buf.Bytes()
	// chop off the newline added by the json encoder
	data = b[:len(b)-1]
	return jsonArrayMimetype, data, nil
}

// copy/sanitize the URL for the Go CDK - it doesn't like params it can't parse
func blobURL(u *url.URL) string {
	out := cloneURL(u)
	q := out.Query()

	for param := range q {
		switch u.Scheme {
		case "s3":
			switch param {
			case "region", "endpoint", "disableSSL", "s3ForcePathStyle":
			default:
				q.Del(param)
			}
		case "gs":
			switch param {
			case "access_id", "private_key_path":
			default:
				q.Del(param)
			}
		}
	}

	if u.Scheme == "s3" {
		// handle AWS_S3_ENDPOINT env var
		endpoint := env.Getenv("AWS_S3_ENDPOINT")
		if endpoint != "" {
			q.Set("endpoint", endpoint)
		}
	}

	out.RawQuery = q.Encode()

	return out.String()
}
