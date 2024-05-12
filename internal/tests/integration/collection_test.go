package integration

import (
	"testing"

	"gotest.tools/v3/fs"
)

func setupCollTest(t *testing.T) *fs.Dir {
	tmpDir := fs.NewDir(t, "gomplate-inttests",
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
	t.Cleanup(tmpDir.Remove)

	return tmpDir
}

func TestColl_Merge(t *testing.T) {
	tmpDir := setupCollTest(t)
	o, e, err := cmd(t,
		"-d", "defaults="+tmpDir.Join("defaults.yaml"),
		"-d", "config="+tmpDir.Join("config.json"),
		"-i", `{{ $defaults := ds "defaults" -}}
		{{ $config := ds "config" -}}
		{{ $merged := coll.Merge $config $defaults -}}
		{{ $merged | data.ToYAML }}`).run()
	assertSuccess(t, o, e, err, `values:
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

func TestColl_Sort(t *testing.T) {
	inOutTest(t, `{{ $maps := jsonArray "[{\"a\": \"foo\", \"b\": 1}, {\"a\": \"bar\", \"b\": 8}, {\"a\": \"baz\", \"b\": 3}]" -}}
{{ range coll.Sort "b" $maps -}}
{{ .a }}
{{ end -}}
`, "foo\nbaz\nbar\n")

	inOutTest(t, `
{{- coll.Sort (coll.Slice "b" "a" "c" "aa") }}
{{ coll.Sort (coll.Slice "b" 14 "c" "aa") }}
{{ coll.Sort (coll.Slice 3.14 3.0 4.0) }}
{{ coll.Sort "Scheme" (coll.Slice (conv.URL "zzz:///") (conv.URL "https:///") (conv.URL "http:///")) }}
`, `[a aa b c]
[b 14 c aa]
[3 3.14 4]
[http:/// https:/// zzz:///]
`)
}

func TestColl_JSONPath(t *testing.T) {
	tmpDir := setupCollTest(t)
	o, e, err := cmd(t, "-c", "config="+tmpDir.Join("config.json"),
		"-i", `{{ .config | jsonpath ".*.three" }}`).run()
	assertSuccess(t, o, e, err, `[5 6 7]`)

	o, e, err = cmd(t, "-c", "config="+tmpDir.Join("config.json"),
		"-i", `{{ .config | coll.JSONPath ".values..a" }}`).run()
	assertSuccess(t, o, e, err, `eh?`)
}

func TestColl_Flatten(t *testing.T) {
	in := "[[1,2],[],[[3,4],[[[5],6],7]]]"
	inOutTest(t, "{{ `"+in+"` | jsonArray | coll.Flatten | toJSON }}", "[1,2,3,4,5,6,7]")
	inOutTest(t, "{{ `"+in+"` | jsonArray | flatten 0 | toJSON }}", in)
	inOutTest(t, "{{ coll.Flatten 1 (`"+in+"` | jsonArray) | toJSON }}", "[1,2,[3,4],[[[5],6],7]]")
	inOutTest(t, "{{ `"+in+"` | jsonArray | coll.Flatten 2 | toJSON }}", "[1,2,3,4,[[5],6],7]")
}

func TestColl_Pick(t *testing.T) {
	inOutTest(t, `{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}{{ coll.Pick "foo" "baz" $data }}`, "map[baz:3 foo:1]")
	inOutTest(t, `{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}{{ coll.Pick (coll.Slice "foo" "baz") $data }}`, "map[baz:3 foo:1]")
}

func TestColl_Omit(t *testing.T) {
	inOutTest(t, `{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}{{ coll.Omit "foo" "baz" $data }}`, "map[bar:2]")
	inOutTest(t, `{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}{{ coll.Omit (coll.Slice "foo" "baz") $data }}`, "map[bar:2]")
}

func TestColl_JQ(t *testing.T) {
	inOutTest(t, `{{ coll.JQ ".foo" (dict "foo" 1 "bar" 2 "baz" 3) }}`, "1")
	inOutTest(t, `{{ coll.Slice "one" 2 "three" 4.0 | jq ".[2]" }}`, `three`)
}
