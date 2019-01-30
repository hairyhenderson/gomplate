//+build integration

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
