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

	"gocloud.dev/blob"

	gaws "github.com/hairyhenderson/gomplate/aws"
	"github.com/hairyhenderson/gomplate/env"
	"github.com/pkg/errors"

	"gocloud.dev/blob/s3blob"
)

func readBlob(source *Source, args ...string) (output []byte, err error) {
	if len(args) >= 2 {
		return nil, errors.New("Maximum two arguments to s3 datasource: alias, extraPath")
	}

	ctx := context.TODO()

	key := source.URL.Path
	if len(args) == 1 {
		key = path.Join(key, args[0])
	}

	// set up a "regular" gomplate AWS SDK session
	sess := gaws.SDKSession()
	// see https://gocloud.dev/concepts/urls/#muxes
	opener := &s3blob.URLOpener{ConfigProvider: sess}
	mux := blob.URLMux{}
	mux.RegisterBucket(s3blob.Scheme, opener)

	bucket, err := mux.OpenBucket(ctx, blobURL(source.URL))
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

func getBlob(ctx context.Context, bucket *blob.Bucket, key string) (mediaType string, data []byte, err error) {
	attr, err := bucket.Attributes(ctx, key)
	if err != nil {
		return "", nil, err
	}
	if attr.ContentType != "" {
		mt, _, e := mime.ParseMediaType(attr.ContentType)
		if e != nil {
			return "", nil, e
		}
		mediaType = mt
	}
	data, err = bucket.ReadAll(ctx, key)
	return mediaType, data, err
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
	out, _ := url.Parse(u.String())
	q := out.Query()
	for param := range q {
		switch param {
		case "region", "endpoint", "disableSSL", "s3ForcePathStyle":
		default:
			q.Del(param)
		}
	}
	// handle AWS_S3_ENDPOINT env var
	endpoint := env.Getenv("AWS_S3_ENDPOINT")
	if endpoint != "" {
		q.Set("endpoint", endpoint)
	}
	out.RawQuery = q.Encode()
	return out.String()
}
