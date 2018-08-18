package integration

import (
	"encoding/json"
	"go/build"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gotestyourself/gotestyourself/icmd"
	vaultapi "github.com/hashicorp/vault/api"
	. "gopkg.in/check.v1"
)

var (
	GomplateBin = build.Default.GOPATH + "/src/github.com/hairyhenderson/gomplate/bin/gomplate"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

// a convenience...
func inOutTest(c *C, i string, o string) {
	result := icmd.RunCommand(GomplateBin, "-i", i)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: o})
}

func handle(c *C, err error) {
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
func waitForURL(c *C, url string) error {
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
