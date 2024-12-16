package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/fs"
)

func setupDatasourcesMergeTest(t *testing.T) (*fs.Dir, *httptest.Server) {
	tmpDir := fs.NewDir(t, "gomplate-inttests",
		fs.WithFiles(map[string]string{
			"config.json": `{"foo": {"bar": "baz"}, "isDefault": false, "isOverride": true}`,
			"default.yml": "foo:\n  bar: qux\nother: true\nisDefault: true\nisOverride: false\n",
		}),
	)
	t.Cleanup(tmpDir.Remove)

	mux := http.NewServeMux()
	mux.HandleFunc("/foo.json", typeHandler("application/json", `{"foo": "bar"}`))
	mux.HandleFunc("/1.env", typeHandler("application/x-env", "FOO=1\nBAR=2\n"))
	mux.HandleFunc("/2.env", typeHandler("application/x-env", "FOO=3\n"))
	// this file is served with a misleading content type and extension, for
	// testing overriding the type
	mux.HandleFunc("/wrongtype.txt", typeHandler("text/html", `{"foo": "bar"}`))
	mux.HandleFunc("/params", paramHandler(t))

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return tmpDir, srv
}

func TestDatasources_Merge(t *testing.T) {
	tmpDir, srv := setupDatasourcesMergeTest(t)

	t.Run("from two aliased datasources", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "user="+tmpDir.Join("config.json"),
			"-d", "default="+tmpDir.Join("default.yml"),
			"-d", "config=merge:user|default",
			"-i", `{{ ds "config" | toJSON }}`,
		).run()
		assertSuccess(t, o, e, err, `{"foo":{"bar":"baz"},"isDefault":false,"isOverride":true,"other":true}`)
	})

	t.Run("with dynamic datasource", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "default="+tmpDir.Join("default.yml"),
			"-d", "config=merge:user|default",
			"-i", `{{ defineDatasource "user" `+"`"+tmpDir.Join("config.json")+"`"+` }}{{ ds "config" | toJSON }}`,
		).run()
		assertSuccess(t, o, e, err, `{"foo":{"bar":"baz"},"isDefault":false,"isOverride":true,"other":true}`)
	})

	t.Run("with inline datasource", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "default="+tmpDir.Join("default.yml"),
			"-d", "config=merge:"+srv.URL+"/foo.json|default",
			"-i", `{{ ds "config" | toJSON }}`,
		).run()
		assertSuccess(t, o, e, err, `{"foo":"bar","isDefault":true,"isOverride":false,"other":true}`)
	})

	t.Run("with two inline env datasources", func(t *testing.T) {
		o, e, err := cmd(t,
			"-c", "merged=merge:"+srv.URL+"/2.env|"+srv.URL+"/1.env",
			"-i", `FOO is {{ .merged.FOO }}`,
		).run()
		assertSuccess(t, o, e, err, `FOO is 3`)
	})

	t.Run("inline merge with inline datasource", func(t *testing.T) {
		o, e, err := cmd(t,
			"-c", "default="+tmpDir.Join("default.yml"),
			"-i", `{{ defineDatasource "merged" "merge:`+srv.URL+`/foo.json|default" -}}
{{ ds "merged" | toJSON }}`,
		).run()
		assertSuccess(t, o, e, err, `{"foo":"bar","isDefault":true,"isOverride":false,"other":true}`)
	})

	t.Run("with overridden type", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "default="+tmpDir.Join("default.yml"),
			"-d", "wrongtype="+srv.URL+"/wrongtype.txt?type=application/json",
			"-d", "config=merge:wrongtype|default",
			"-i", `{{ ds "config" | toJSON }}`,
		).run()
		assertSuccess(t, o, e, err, `{"foo":"bar","isDefault":true,"isOverride":false,"other":true}`)
	})

	t.Run("type overridden by env var", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "default="+tmpDir.Join("default.yml"),
			"-d", "wrongtype="+srv.URL+"/wrongtype.txt?_=application/json",
			"-d", "config=merge:wrongtype|default",
			"-i", `{{ ds "config" | toJSON }}`,
		).withEnv("GOMPLATE_TYPE_PARAM", "_").run()
		assertSuccess(t, o, e, err, `{"foo":"bar","isDefault":true,"isOverride":false,"other":true}`)

		o, e, err = cmd(t,
			"-c", "default="+tmpDir.Join("default.yml"),
			"-c", "params="+srv.URL+"/params?foo=bar&type=http&_type=application/json",
			"-c", "merged=merge:params|default",
			"-i", `{{ .merged | toJSON }}`,
		).withEnv("GOMPLATE_TYPE_PARAM", "_type").run()
		assertSuccess(t, o, e, err, `{"foo":["bar"],"isDefault":true,"isOverride":false,"other":true,"type":["http"]}`)
	})

	t.Run("from stdin", func(t *testing.T) {
		o, e, err := cmd(t,
			"-d", "stdindata=stdin:///in.json",
			"-d", "filedata="+srv.URL+"/foo.json",
			"-d", "merged=merge:stdindata|filedata",
			"-i", `{{ ds "merged" | toJSON }}`,
		).withStdin(`{"baz": "qux"}`).run()
		assertSuccess(t, o, e, err, `{"baz":"qux","foo":"bar"}`)
	})
}
