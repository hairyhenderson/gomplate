package integration

import (
	. "gopkg.in/check.v1"

	"gotest.tools/v3/fs"
)

type CollSuite struct {
	tmpDir *fs.Dir
}

var _ = Suite(&CollSuite{})

func (s *CollSuite) SetUpTest(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFiles(map[string]string{
			"defaults.yaml": `values:
  one: 1
  two: 2
  three:
    - 4
  four:
    a: a
    b: b
`,
			"config.json": `{
				"values": {
					"one": "uno",
					"three": [ 5, 6, 7 ],
					"four": { "a": "eh?" }
				}
			}`,
		}))
}

func (s *CollSuite) TearDownTest(c *C) {
	s.tmpDir.Remove()
}

func (s *CollSuite) TestCollMerge(c *C) {
	o, e, err := cmdTest(c,
		"-d", "defaults="+s.tmpDir.Join("defaults.yaml"),
		"-d", "config="+s.tmpDir.Join("config.json"),
		"-i", `{{ $defaults := ds "defaults" -}}
		{{ $config := ds "config" -}}
		{{ $merged := coll.Merge $config $defaults -}}
		{{ $merged | data.ToYAML }}`)
	assertSuccess(c, o, e, err, `values:
  four:
    a: eh?
    b: b
  one: uno
  three:
    - 5
    - 6
    - 7
  two: 2
`)
}

func (s *CollSuite) TestSort(c *C) {
	inOutTest(c, `{{ $maps := jsonArray "[{\"a\": \"foo\", \"b\": 1}, {\"a\": \"bar\", \"b\": 8}, {\"a\": \"baz\", \"b\": 3}]" -}}
{{ range coll.Sort "b" $maps -}}
{{ .a }}
{{ end -}}
`, "foo\nbaz\nbar\n")

	inOutTest(c, `
{{- coll.Sort (slice "b" "a" "c" "aa") }}
{{ coll.Sort (slice "b" 14 "c" "aa") }}
{{ coll.Sort (slice 3.14 3.0 4.0) }}
{{ coll.Sort "Scheme" (coll.Slice (conv.URL "zzz:///") (conv.URL "https:///") (conv.URL "http:///")) }}
`, `[a aa b c]
[b 14 c aa]
[3 3.14 4]
[http:/// https:/// zzz:///]
`)
}

func (s *CollSuite) TestJSONPath(c *C) {
	o, e, err := cmdTest(c, "-c", "config="+s.tmpDir.Join("config.json"),
		"-i", `{{ .config | jsonpath ".*.three" }}`)
	assertSuccess(c, o, e, err, `[5 6 7]`)

	o, e, err = cmdTest(c, "-c", "config="+s.tmpDir.Join("config.json"),
		"-i", `{{ .config | coll.JSONPath ".values..a" }}`)
	assertSuccess(c, o, e, err, `eh?`)
}

func (s *CollSuite) TestFlatten(c *C) {
	in := "[[1,2],[],[[3,4],[[[5],6],7]]]"
	inOutTest(c, "{{ `"+in+"` | jsonArray | coll.Flatten | toJSON }}", "[1,2,3,4,5,6,7]")
	inOutTest(c, "{{ `"+in+"` | jsonArray | flatten 0 | toJSON }}", in)
	inOutTest(c, "{{ coll.Flatten 1 (`"+in+"` | jsonArray) | toJSON }}", "[1,2,[3,4],[[[5],6],7]]")
	inOutTest(c, "{{ `"+in+"` | jsonArray | coll.Flatten 2 | toJSON }}", "[1,2,3,4,[[5],6],7]")
}

func (s *CollSuite) TestPick(c *C) {
	inOutTest(c, `{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}{{ coll.Pick "foo" "baz" $data }}`, "map[baz:3 foo:1]")
}

func (s *CollSuite) TestOmit(c *C) {
	inOutTest(c, `{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}{{ coll.Omit "foo" "baz" $data }}`, "map[bar:2]")
}
