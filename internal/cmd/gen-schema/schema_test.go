package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/hairyhenderson/yaml"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func loadSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()

	f, err := os.Open("../../../schema/gomplate-config.json")
	if err != nil {
		t.Fatalf("open schema: %v", err)
	}
	defer f.Close()

	doc, err := jsonschema.UnmarshalJSON(f)
	if err != nil {
		t.Fatalf("parse schema JSON: %v", err)
	}

	c := jsonschema.NewCompiler()
	if err = c.AddResource("gomplate-config.json", doc); err != nil {
		t.Fatalf("add schema resource: %v", err)
	}

	sch, err := c.Compile("gomplate-config.json")
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}
	return sch
}

// yamlToAny converts a YAML string to a value suitable for schema validation
// by round-tripping through JSON (normalises types to what the schema expects).
func yamlToAny(t *testing.T, src string) any {
	t.Helper()

	var raw any
	if err := yaml.Unmarshal([]byte(src), &raw); err != nil {
		t.Fatalf("unmarshal YAML: %v", err)
	}

	// Round-trip through JSON so numeric types match JSON Schema expectations.
	b, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal to JSON: %v", err)
	}

	var doc any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	return doc
}

func TestSchema(t *testing.T) {
	sch := loadSchema(t)

	cases := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		// --- valid configs ---
		{
			name: "empty config",
			yaml: `{}`,
		},
		{
			name: "simple inputDir/outputDir",
			yaml: `
inputDir: templates/
outputDir: out/
`,
		},
		{
			name: "inputFiles and outputFiles",
			yaml: `
inputFiles: [first.tmpl, second.tmpl]
outputFiles: [first.out, second.out]
`,
		},
		{
			name: "inline input",
			yaml: `
in: hello to {{ env.Env.USER }}
outputFiles: [out.txt]
`,
		},
		{
			name: "datasource with url only",
			yaml: `
datasources:
  stuff:
    url: stuff.yaml
`,
		},
		{
			name: "datasource with url and headers",
			yaml: `
datasources:
  remote:
    url: https://example.com/api/v1/data
    header:
      Authorization: ["Basic aGF4MHI6c3dvcmRmaXNoCg=="]
`,
		},
		{
			name:    "datasource as bare string (removed in v5)",
			yaml:    `datasources: {local: "file:///tmp/data.json"}`,
			wantErr: true,
		},
		{
			name: "context sources",
			yaml: `
context:
  data:
    url: https://example.com/api/v1/data
    header:
      Authorization: ["Basic aGF4MHI6c3dvcmRmaXNoCg=="]
  stuff:
    url: stuff.yaml
`,
		},
		{
			name: "context dot alias",
			yaml: `
context:
  .:
    url: data.toml
`,
		},
		{
			name: "templates with url",
			yaml: `
templates:
  t:
    url: file:///foo/bar/helloworld.tmpl
  remote:
    url: https://example.com/sometemplate
    header:
      Authorization: ["Basic aGF4MHI6c3dvcmRmaXNoCg=="]
`,
		},
		{
			name:    "template as bare string (removed in v5)",
			yaml:    `templates: {dir: "/foo/bar"}`,
			wantErr: true,
		},
		{
			name: "plugin as bare string",
			yaml: `
plugins:
  figlet: /usr/local/bin/figlet
`,
		},
		{
			name: "plugin as map",
			yaml: `
plugins:
  figlet:
    cmd: /usr/local/bin/figlet
    args: [oh, hello]
    pipe: true
    timeout: 1s
`,
		},
		{
			name: "pluginTimeout duration",
			yaml: `
plugins:
  figlet: /usr/local/bin/figlet
pluginTimeout: 500ms
`,
		},
		{
			name: "delimiters",
			yaml: `
leftDelim: '(('
rightDelim: '))'
`,
		},
		{
			name: "missingKey",
			yaml: `
missingKey: error
`,
		},
		{
			name: "chmod",
			yaml: `
chmod: "644"
`,
		},
		{
			name: "excludes and excludeProcessing",
			yaml: `
inputDir: in/
outputDir: out/
excludes: ['*.txt', '!keep.txt']
excludeProcessing: ['*.jpg']
`,
		},
		{
			name: "postExec and execPipe",
			yaml: `
in: hello
postExec: [cat]
execPipe: true
`,
		},
		{
			name: "experimental",
			yaml: `
experimental: true
`,
		},
		{
			name: "outputMap",
			yaml: `
inputDir: in/
outputMap: 'out/{{ .in }}'
`,
		},

		// --- invalid configs ---
		{
			name:    "unknown field",
			yaml:    `inputdirr: typo/`,
			wantErr: true,
		},
		{
			name:    "execPipe wrong type",
			yaml:    `execPipe: "yes"`,
			wantErr: true,
		},
		{
			name:    "experimental wrong type",
			yaml:    `experimental: 1`,
			wantErr: true,
		},
		{
			name:    "pluginTimeout wrong type (integer)",
			yaml:    `pluginTimeout: 500`,
			wantErr: true,
		},
		{
			name:    "inputFiles wrong type (string instead of array)",
			yaml:    `inputFiles: first.tmpl`,
			wantErr: true,
		},
		{
			name:    "datasource wrong type (integer)",
			yaml:    `datasources: {data: 42}`,
			wantErr: true,
		},
		{
			name:    "missingKey invalid value",
			yaml:    `missingKey: errror`,
			wantErr: true,
		},
		{
			name:    "datasource unknown field",
			yaml:    "datasources:\n  d:\n    urll: foo\n",
			wantErr: true,
		},
		{
			name:    "plugin unknown field",
			yaml:    "plugins:\n  p:\n    cmd: /bin/foo\n    timout: 1s\n",
			wantErr: true,
		},
		{
			name:    "plugin map missing cmd",
			yaml:    "plugins:\n  p:\n    pipe: true\n",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := yamlToAny(t, tc.yaml)
			err := sch.Validate(doc)
			if tc.wantErr && err == nil {
				t.Error("expected validation error, got none")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}
