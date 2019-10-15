//+build integration

package integration

import (
	"bytes"
	"net"
	"net/http"
	"os"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	. "gopkg.in/check.v1"

	"gotest.tools/v3/icmd"
)

type BlobDatasourcesSuite struct {
	l *net.TCPListener
}

var _ = Suite(&BlobDatasourcesSuite{})

func (s *BlobDatasourcesSuite) SetUpSuite(c *C) {
	backend := s3mem.New()
	s3 := gofakes3.New(backend)
	var err error
	s.l, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1")})
	handle(c, err)

	http.Handle("/", s3.Server())
	go http.Serve(s.l, nil)

	err = backend.CreateBucket("mybucket")
	handle(c, err)
	contents := `{"value":"json", "name":"foo"}`
	_, err = backend.PutObject("mybucket", "foo.json", map[string]string{"Content-Type": "application/json"}, bytes.NewBufferString(contents), int64(len(contents)))
	handle(c, err)

	contents = `{"value":"json", "name":"bar"}`
	_, err = backend.PutObject("mybucket", "bar.json", map[string]string{"Content-Type": "application/json"}, bytes.NewBufferString(contents), int64(len(contents)))
	handle(c, err)

	contents = `hello world`
	_, err = backend.PutObject("mybucket", "a/b/c/hello.txt", map[string]string{"Content-Type": "text/plain"}, bytes.NewBufferString(contents), int64(len(contents)))
	handle(c, err)

	contents = `goodbye world`
	_, err = backend.PutObject("mybucket", "a/b/c/goodbye.txt", map[string]string{"Content-Type": "text/plain"}, bytes.NewBufferString(contents), int64(len(contents)))
	handle(c, err)

	contents = "a: foo\nb: bar\nc:\n  cc: yay for yaml\n"
	_, err = backend.PutObject("mybucket", "a/b/c/d", map[string]string{"Content-Type": "application/yaml"}, bytes.NewBufferString(contents), int64(len(contents)))
	handle(c, err)
}

func (s *BlobDatasourcesSuite) TearDownSuite(c *C) {
	s.l.Close()
}

func (s *BlobDatasourcesSuite) TestS3Datasource(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-c", "data=s3://ryft-public-sample-data/integration_test_dataset.json?region=us-east-1&type=application/array+json",
		"-i", "{{ $d := index .data 0 }}{{ $d.firstName }} {{ $d.lastName }}",
	), func(c *icmd.Cmd) {
		c.Env = []string{
			"AWS_ANON=true",
			"AWS_TIMEOUT=5000",
		}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "Petra Mcintyre"})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-c", "data=s3://mybucket/foo.json?"+
			"region=us-east-1&"+
			"disableSSL=true&"+
			"endpoint="+s.l.Addr().String()+"&"+
			"s3ForcePathStyle=true",
		"-i", "{{ .data.value }}",
	), func(c *icmd.Cmd) {
		c.Env = []string{
			"AWS_ACCESS_KEY_ID=YOUR-ACCESSKEYID",
			"AWS_SECRET_ACCESS_KEY=YOUR-SECRETACCESSKEY",
		}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "json"})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-c", "data=s3://mybucket/foo.json?"+
			"region=us-east-1&"+
			"disableSSL=true&"+
			"s3ForcePathStyle=true",
		"-i", "{{ .data.value }}",
	), func(c *icmd.Cmd) {
		c.Env = []string{
			"AWS_ANON=true",
			"AWS_S3_ENDPOINT=" + s.l.Addr().String(),
		}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "json"})
}

func (s *BlobDatasourcesSuite) TestS3Directory(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-c", "data=s3://ryft-public-sample-data/?region=us-east-1",
		"-i", "{{ index .data 0 }}",
	), func(c *icmd.Cmd) {
		c.Env = []string{
			"AWS_ANON=true",
			"AWS_TIMEOUT=15000",
		}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "AWS-x86-AMI-queries.json"})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-c", "data=s3://mybucket/a/b/c/?"+
			"region=us-east-1&"+
			"disableSSL=true&"+
			"endpoint="+s.l.Addr().String()+"&"+
			"s3ForcePathStyle=true",
		"-i", "{{ .data }}",
	), func(c *icmd.Cmd) {
		c.Env = []string{
			"AWS_ACCESS_KEY_ID=YOUR-ACCESSKEYID",
			"AWS_SECRET_ACCESS_KEY=YOUR-SECRETACCESSKEY",
		}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "[d goodbye.txt hello.txt]"})
}

func (s *BlobDatasourcesSuite) TestS3MIMETypes(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-c", "data=s3://mybucket/a/b/c/d?"+
			"region=us-east-1&"+
			"disableSSL=true&"+
			"endpoint="+s.l.Addr().String()+"&"+
			"s3ForcePathStyle=true",
		"-i", "{{ .data.c.cc }}",
	), func(c *icmd.Cmd) {
		c.Env = []string{"AWS_ANON=true"}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "yay for yaml"})
}

func (s *BlobDatasourcesSuite) TestGCSDatasource(c *C) {
	// this only works if we're authed with GCS
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		c.Skip("Not configured to authenticate with Google Cloud - skipping")
		return
	}
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-c", "data=gs://gcp-public-data-landsat/LT08/PRE/015/013/LT80150132013127LGN01/LT80150132013127LGN01_MTL.txt?type=text/plain",
		"-i", "{{ len .data }}",
	), func(c *icmd.Cmd) {
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "3218"})
}

func (s *BlobDatasourcesSuite) TestGCSDirectory(c *C) {
	// this only works if we're likely to be authed with GCS
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		c.Skip("Not configured to authenticate with Google Cloud - skipping")
		return
	}
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-c", "data=gs://gcp-public-data-landsat/",
		"-i", "{{ coll.Has .data `index.csv.gz` }}",
	), func(c *icmd.Cmd) {})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "true"})
}
