package integration

import (
	"bytes"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/require"
)

func setupDatasourcesBlobTest(t *testing.T) *httptest.Server {
	backend := s3mem.New()
	s3 := gofakes3.New(backend)

	srv := httptest.NewServer(s3.Server())
	t.Cleanup(srv.Close)

	err := backend.CreateBucket("mybucket")
	require.NoError(t, err)
	contents := `{"value":"json", "name":"foo"}`
	_, err = backend.PutObject("mybucket", "foo.json", map[string]string{"Content-Type": "application/json"}, bytes.NewBufferString(contents), int64(len(contents)))
	require.NoError(t, err)

	contents = `{"value":"json", "name":"bar"}`
	_, err = backend.PutObject("mybucket", "bar.json", map[string]string{"Content-Type": "application/json"}, bytes.NewBufferString(contents), int64(len(contents)))
	require.NoError(t, err)

	contents = `hello world`
	_, err = backend.PutObject("mybucket", "a/b/c/hello.txt", map[string]string{"Content-Type": "text/plain"}, bytes.NewBufferString(contents), int64(len(contents)))
	require.NoError(t, err)

	contents = `goodbye world`
	_, err = backend.PutObject("mybucket", "a/b/c/goodbye.txt", map[string]string{"Content-Type": "text/plain"}, bytes.NewBufferString(contents), int64(len(contents)))
	require.NoError(t, err)

	contents = "a: foo\nb: bar\nc:\n  cc: yay for yaml\n"
	_, err = backend.PutObject("mybucket", "a/b/c/d", map[string]string{"Content-Type": "application/yaml"}, bytes.NewBufferString(contents), int64(len(contents)))
	require.NoError(t, err)

	return srv
}

func TestDatasources_Blob_S3Datasource(t *testing.T) {
	o, e, err := cmd(t,
		"-c", "data=s3://noaa-bathymetry-pds/csv/2022/03/02/20220302_056e577c7cd8323fdd8a04d3812cf78e_pointData.csv?region=us-east-1&type=text/csv",
		"-i", `{{ index (index .data 0) 6 }}: {{ index (index .data 1) 6 }}
{{ index (index .data 0) 5 }}: {{ index (index .data 1) 5 }}`).
		withEnv("AWS_ANON", "true").
		withEnv("AWS_TIMEOUT", "5000").
		run()
	assertSuccess(t, o, e, err, `PLATFORM_NAME: Ramform Hyperion
TIME: 2022-03-01T22:00:04.000Z`)

	srv := setupDatasourcesBlobTest(t)

	o, e, err = cmd(t,
		"-c", "data=s3://mybucket/foo.json?"+
			"region=us-east-1&"+
			"disableSSL=true&"+
			"endpoint="+srv.Listener.Addr().String()+"&"+
			"s3ForcePathStyle=true",
		"-i", "{{ .data.value }}").
		withEnv("AWS_ACCESS_KEY_ID", "YOUR-ACCESSKEYID").
		withEnv("AWS_SECRET_ACCESS_KEY", "YOUR-SECRETACCESSKEY").
		run()
	assertSuccess(t, o, e, err, "json")

	o, e, err = cmd(t,
		"-c", "data=s3://mybucket/foo.json?"+
			"region=us-east-1&"+
			"disableSSL=true&"+
			"s3ForcePathStyle=true",
		"-i", "{{ .data.value }}").
		withEnv("AWS_ANON", "true").
		withEnv("AWS_S3_ENDPOINT", srv.Listener.Addr().String()).
		run()
	assertSuccess(t, o, e, err, "json")
}

func TestDatasources_Blob_S3Directory(t *testing.T) {
	// This recently replaced ryft-public-sample-data after access was disabled.
	// This bucket came from https://registry.opendata.aws, and is public. The
	// content isn't important, just that it's something we can read and parse
	// on a _real_ S3 bucket.
	o, e, err := cmd(t, "-c", "data=s3://noaa-bathymetry-pds/csv/2022/03/02/?region=us-east-1",
		"-i", "{{ index .data 0 }}").
		withEnv("AWS_ANON", "true").
		withEnv("AWS_TIMEOUT", "15000").
		run()
	assertSuccess(t, o, e, err, "20220302_056e577c7cd8323fdd8a04d3812cf78e_pointData.csv")

	srv := setupDatasourcesBlobTest(t)

	o, e, err = cmd(t, "-c", "data=s3://mybucket/a/b/c/?"+
		"region=us-east-1&"+
		"disableSSL=true&"+
		"endpoint="+srv.Listener.Addr().String()+"&"+
		"s3ForcePathStyle=true",
		"-i", "{{ .data }}").
		withEnv("AWS_ACCESS_KEY_ID", "YOUR-ACCESSKEYID").
		withEnv("AWS_SECRET_ACCESS_KEY", "YOUR-SECRETACCESSKEY").
		run()
	assertSuccess(t, o, e, err, "[d goodbye.txt hello.txt]")
}

func TestDatasources_Blob_S3MIMETypes(t *testing.T) {
	srv := setupDatasourcesBlobTest(t)
	o, e, err := cmd(t, "-c", "data=s3://mybucket/a/b/c/d?"+
		"region=us-east-1&"+
		"disableSSL=true&"+
		"endpoint="+srv.Listener.Addr().String()+"&"+
		"s3ForcePathStyle=true",
		"-i", "{{ .data.c.cc }}").
		withEnv("AWS_ANON", "true").run()
	assertSuccess(t, o, e, err, "yay for yaml")
}

func TestDatasources_Blob_GCSDatasource(t *testing.T) {
	// this only works if we're authed with GCS
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		t.Skip("Not configured to authenticate with Google Cloud - skipping")
		return
	}

	o, e, err := cmd(t, "-c", "data=gs://gcp-public-data-landsat/LT08/01/015/013/LT08_L1GT_015013_20130315_20170310_01_T2/LT08_L1GT_015013_20130315_20170310_01_T2_MTL.txt?type=text/plain",
		"-i", "{{ len .data }}").
		withEnv("GOOGLE_APPLICATION_CREDENTIALS",
			os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")).run()
	assertSuccess(t, o, e, err, "3672")
}

func TestDatasources_Blob_GCSDirectory(t *testing.T) {
	// this only works if we're likely to be authed with GCS
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		t.Skip("Not configured to authenticate with Google Cloud - skipping")
		return
	}

	o, e, err := cmd(t, "-c", "data=gs://gcp-public-data-landsat/",
		"-i", "{{ coll.Has .data `index.csv.gz` }}").
		withEnv("GOOGLE_APPLICATION_CREDENTIALS",
			os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")).run()
	assertSuccess(t, o, e, err, "true")
}
