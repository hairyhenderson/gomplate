package integration

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"runtime"
	"testing"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

const isWindows = runtime.GOOS == "windows"

// a convenience...
func inOutTest(t *testing.T, i, o string) {
	t.Helper()

	stdout, stderr, err := cmd(t, "-i", i).run()
	assert.NoError(t, err)
	assert.Equal(t, "", stderr)
	assert.Equal(t, o, stdout)
}

func inOutContains(t *testing.T, i, o string) {
	t.Helper()

	stdout, stderr, err := cmd(t, "-i", i).run()
	assert.NoError(t, err)
	assert.Equal(t, "", stderr)
	assert.Contains(t, stdout, o)
}

func assertSuccess(t *testing.T, o, e string, err error, expected string) {
	t.Helper()

	assert.NoError(t, err)
	assert.Equal(t, "", e)
	assert.Equal(t, expected, o)
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
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", t)
		w.Write([]byte(body))
	}
}

// freeport - find a free TCP port for immediate use. No guarantees!
func freeport() (port int, addr string) {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	a := l.Addr().(*net.TCPAddr)
	port = a.Port
	return port, a.String()
}

// waitForURL - waits up to 20s for a given URL to respond with a 200
func waitForURL(t *testing.T, url string) error {
	client := http.DefaultClient
	retries := 100
	for retries > 0 {
		retries--
		time.Sleep(200 * time.Millisecond)
		resp, err := client.Get(url)
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
	return nil
}

type vaultClient struct {
	addr      string
	rootToken string
	vc        *vaultapi.Client
}

func createVaultClient(addr string, rootToken string) (*vaultClient, error) {
	config := vaultapi.DefaultConfig()
	config.Address = "http://" + addr
	client, err := vaultapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	v := &vaultClient{
		addr:      addr,
		rootToken: rootToken,
		vc:        client,
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
	c.env[k] = v
	return c
}
