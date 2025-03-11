package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	gcmd "github.com/hairyhenderson/gomplate/v4/internal/cmd"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/icmd"
)

const isWindows = runtime.GOOS == "windows"

// a convenience...
func inOutTest(t *testing.T, i, o string, args ...string) {
	t.Helper()

	args = append(args, "-i", i)
	stdout, stderr, err := cmd(t, args...).run()
	assertSuccess(t, stdout, stderr, err, o)
}

func inOutContainsError(t *testing.T, i, e string, args ...string) {
	t.Helper()

	args = append(args, "-i", i)
	stdout, stderr, err := cmd(t, args...).run()
	assertFailed(t, stdout, stderr, err, e)
}

func inOutTestExperimental(t *testing.T, i, o string) {
	t.Helper()

	stdout, stderr, err := cmd(t, "--experimental", "-i", i).run()
	assertSuccess(t, stdout, stderr, err, o)
}

func inOutContains(t *testing.T, i, o string) {
	t.Helper()

	stdout, stderr, err := cmd(t, "-i", i).run()
	assert.Equal(t, "", stderr)
	assert.Contains(t, stdout, o)
	require.NoError(t, err)
}

func assertSuccess(t *testing.T, o, e string, err error, expected string) {
	t.Helper()

	require.NoError(t, err)
	// Filter out AWS SDK checksum warnings
	filteredErr := filterAWSChecksumWarnings(e)
	assert.Equal(t, "", filteredErr)
	assert.Equal(t, expected, o)
}

// filterAWSChecksumWarnings removes AWS SDK checksum warning messages from the
// error output. These are a non-issue for our tests, since we use gofakes3 and
// anonymous buckets.
func filterAWSChecksumWarnings(e string) string {
	lines := strings.Split(e, "\n")

	filteredLines := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.Contains(line, "WARN Response has no supported checksum. Not validating response payload.") {
			filteredLines = append(filteredLines, line)
		}
	}

	return strings.Join(filteredLines, "\n")
}

func assertFailed(t *testing.T, o, e string, err error, expected string) {
	t.Helper()

	assert.Contains(t, e, expected)
	assert.Equal(t, "", o)
	require.Error(t, err)
}

// mirrorHandler - reflects back the HTTP headers from the request
func mirrorHandler(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		Headers http.Header `json:"headers"`
	}
	req := Req{r.Header}
	b, err := json.Marshal(req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func typeHandler(t, body string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", t)
		w.Write([]byte(body))
	}
}

func paramHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// just returns params as JSON
		w.Header().Set("Content-Type", "application/json")

		enc := json.NewEncoder(w)
		if err := enc.Encode(r.URL.Query()); err != nil {
			t.Fatalf("error encoding: %v", err)
		}
	}
}

// freeport - find a free TCP port for immediate use. No guarantees!
func freeport(t *testing.T) (port int, addr string) {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	a := l.Addr().(*net.TCPAddr)
	port = a.Port
	return port, a.String()
}

// waitForURL - waits up to 20s for a given URL to respond with a 200
func waitForURL(t *testing.T, url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client := http.DefaultClient
	retries := 100
	for retries > 0 {
		retries--
		time.Sleep(200 * time.Millisecond)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Got error, retries left: %d (error: %v)", retries, err)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		t.Logf("Body is: %s", body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			return nil
		}
	}
	return fmt.Errorf("URL %s never responded with 200", url)
}

type vaultClient struct {
	vc        *vaultapi.Client
	addr      string
	rootToken string
}

func createVaultClient(addr string, rootToken string) (*vaultClient, error) {
	config := vaultapi.DefaultConfig()
	config.Address = "http://" + addr
	client, err := vaultapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	v := &vaultClient{
		vc:        client,
		addr:      addr,
		rootToken: rootToken,
	}
	client.SetToken(rootToken)
	return v, nil
}

func (v *vaultClient) tokenCreate(policy string, uses int) (string, error) {
	opts := &vaultapi.TokenCreateRequest{
		Policies: []string{policy},
		TTL:      "1m",
		NumUses:  uses,
	}
	token, err := v.vc.Auth().Token().Create(opts)
	if err != nil {
		return "", err
	}
	return token.Auth.ClientToken, nil
}

type command struct {
	t     *testing.T
	dir   string
	stdin string
	env   map[string]string
	envK  []string
	args  []string
}

func cmd(t *testing.T, args ...string) *command {
	return &command{t: t, args: args}
}

func (c *command) withDir(dir string) *command {
	c.dir = dir
	return c
}

func (c *command) withStdin(in string) *command {
	c.stdin = in
	return c
}

func (c *command) withEnv(k, v string) *command {
	if c.env == nil {
		c.env = map[string]string{}
	}
	if c.envK == nil {
		c.envK = []string{}
	}
	c.env[k] = v
	c.envK = append(c.envK, k)
	return c
}

// set this at 'go test' time to test with a pre-compiled binary instead of
// running all tests in-process
var GomplateBinPath = ""

func (c *command) run() (o, e string, err error) {
	c.t.Helper()

	if GomplateBinPath != "" {
		return c.runCompiled(GomplateBinPath)
	}
	return c.runInProcess()
}

func (c *command) runInProcess() (o, e string, err error) {
	c.t.Helper()

	// iterate env vars by order of insertion
	for _, k := range c.envK {
		// clean up after ourselves
		if orig, ok := os.LookupEnv(k); ok {
			defer os.Setenv(k, orig)
		} else {
			defer os.Unsetenv(k)
		}
		os.Setenv(k, c.env[k])
	}

	if c.dir != "" {
		//nolint:govet
		origWd, err := os.Getwd()
		if err != nil {
			c.t.Fatal(err)
		}
		defer os.Chdir(origWd)

		err = os.Chdir(c.dir)
		if err != nil {
			c.t.Fatal(err)
		}

		c.t.Logf("running in dir %q", c.dir)
	}

	stdin := strings.NewReader(c.stdin)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	err = gcmd.Main(ctx, c.args, stdin, stdout, stderr)
	return stdout.String(), stderr.String(), err
}

func (c *command) runCompiled(bin string) (o, e string, err error) {
	cmd := icmd.Command(bin, c.args...)
	cmd.Dir = c.dir
	cmd.Stdin = strings.NewReader(c.stdin)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOMPLATE_LOG_FORMAT=simple")
	for _, k := range c.envK {
		cmd.Env = append(cmd.Env, k+"="+c.env[k])
	}

	result := icmd.RunCmd(cmd)
	if result.Error != nil {
		result.Error = fmt.Errorf("%w: %s", result.Error, result.Stderr())
	}
	return result.Stdout(), result.Stderr(), result.Error
}
