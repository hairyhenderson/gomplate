package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hairyhenderson/gomplate/v3/internal/cmd"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"gopkg.in/check.v1"
)

const isWindows = runtime.GOOS == "windows"

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	check.TestingT(t)
}

// a convenience...
func inOutTest(c *check.C, i, o string) {
	stdout, stderr, err := cmdTest(c, "-i", i)
	assert.NoError(c, err)
	assert.Equal(c, "", stderr)
	assert.Equal(c, o, stdout)
}

func inOutContains(c *check.C, i, o string) {
	stdout, stderr, err := cmdTest(c, "-i", i)
	assert.NoError(c, err)
	assert.Equal(c, "", stderr)
	assert.Contains(c, stdout, o)
}

func cmdTest(c *check.C, args ...string) (o, e string, err error) {
	ctx := context.Background()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	err = cmd.Main(ctx, args, nil, stdout, stderr)
	return stdout.String(), stderr.String(), err
}

func cmdWithStdin(c *check.C, args []string, in string) (o, e string, err error) {
	ctx := context.Background()
	stdin := strings.NewReader(in)
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	err = cmd.Main(ctx, args, stdin, stdout, stderr)
	return stdout.String(), stderr.String(), err
}

func cmdWithDir(c *check.C, dir string, args ...string) (o, e string, err error) {
	origWd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	defer func() { os.Chdir(origWd) }()
	err = os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	return cmdTest(c, args...)
}

func cmdWithEnv(c *check.C, args []string, env map[string]string) (o, e string, err error) {
	origEnviron := map[string]string{}
	for k, v := range env {
		origEnviron[k] = os.Getenv(k)
		os.Setenv(k, v)
	}

	defer func() {
		for k, v := range origEnviron {
			os.Setenv(k, v)
		}
	}()

	ctx := context.Background()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	err = cmd.Main(ctx, args, nil, stdout, stderr)
	return stdout.String(), stderr.String(), err
}

func assertSuccess(c *check.C, o, e string, err error, expected string) {
	assert.NoError(c, err)
	assert.Equal(c, "", e)
	assert.Equal(c, expected, o)
}

func handle(c *check.C, err error) {
	if err != nil {
		c.Fatal(err)
	}
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
func waitForURL(c *check.C, url string) error {
	client := http.DefaultClient
	retries := 100
	for retries > 0 {
		retries--
		time.Sleep(200 * time.Millisecond)
		resp, err := client.Get(url)
		if err != nil {
			c.Logf("Got error, retries left: %d (error: %v)", retries, err)
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		c.Logf("Body is: %s", body)
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

func killByPidFile(pidFile string) error {
	p, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(p)))
	if err != nil {
		return err
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	err = process.Kill()
	return err
}
