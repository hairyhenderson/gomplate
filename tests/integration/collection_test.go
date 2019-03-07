//+build integration
//+build !windows

package integration

import (
	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/fs"
	"github.com/gotestyourself/gotestyourself/icmd"
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

func (s *CollSuite) TestMerge(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "defaults="+s.tmpDir.Join("defaults.yaml"),
		"-d", "config="+s.tmpDir.Join("config.json"),
		"-i", `{{ $defaults := ds "defaults" -}}
		{{ $config := ds "config" -}}
		{{ $merged := coll.Merge $config $defaults -}}
		{{ $merged | data.ToYAML }}
`))
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `values:
  four:
    a: eh?
    b: b
  one: uno
  three:
  - 5
  - 6
  - 7
  two: 2
`})
}

func (s *CollSuite) TestSort(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-i", `{{ $maps := jsonArray "[{\"a\": \"foo\", \"b\": 1}, {\"a\": \"bar\", \"b\": 8}, {\"a\": \"baz\", \"b\": 3}]" -}}
{{ range coll.Sort "b" $maps -}}
{{ .a }}
{{ end -}}
`))
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "foo\nbaz\nbar\n"})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-i", `
{{- coll.Sort (slice "b" "a" "c" "aa") }}
{{ coll.Sort (slice "b" 14 "c" "aa") }}
{{ coll.Sort (slice 3.14 3.0 4.0) }}
{{ coll.Sort "Scheme" (coll.Slice (conv.URL "zzz:///") (conv.URL "https:///") (conv.URL "http:///")) }}
`))
	result.Assert(c, icmd.Expected{ExitCode: 0,
		Out: `[a aa b c]
[b 14 c aa]
[3 3.14 4]
[http:/// https:/// zzz:///]
`,
	})
}
