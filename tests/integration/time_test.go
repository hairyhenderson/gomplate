//+build integration

package integration

import (
	"runtime"
	"time"

	. "gopkg.in/check.v1"
	"gotest.tools/v3/icmd"
)

type TimeSuite struct{}

var _ = Suite(&TimeSuite{})

func (s *TimeSuite) TestTime(c *C) {
	f := `Mon Jan 02 15:04:05 MST 2006`
	i := `Fri Feb 13 23:31:30 UTC 2009`
	inOutTest(c, `{{ (time.Parse "`+f+`" "`+i+`").Format "2006-01-02 15 -0700" }}`,
		"2009-02-13 23 +0000")

	if runtime.GOOS != "windows" {
		result := icmd.RunCmd(icmd.Command(GomplateBin, "-i",
			`{{ (time.ParseLocal time.Kitchen "6:00AM").Format "15:04 MST" }}`), func(cmd *icmd.Cmd) {
			cmd.Env = []string{"TZ=Africa/Luanda"}
		})
		result.Assert(c, icmd.Expected{ExitCode: 0, Out: "06:00 LMT"})

		result = icmd.RunCmd(icmd.Command(GomplateBin, "-i",
			`{{ time.ZoneOffset }}`), func(cmd *icmd.Cmd) {
			cmd.Env = []string{"TZ=UTC"}
		})
		result.Assert(c, icmd.Expected{ExitCode: 0, Out: "0"})
	}

	zname, _ := time.Now().Zone()
	inOutTest(c, `{{ time.ZoneName }}`, zname)

	inOutTest(c, `{{ (time.Now).Format "2006-01-02 15 -0700" }}`,
		time.Now().Format("2006-01-02 15 -0700"))

	inOutTest(c, `{{ (time.ParseInLocation time.Kitchen "Africa/Luanda" "6:00AM").Format "15:04 MST" }}`,
		"06:00 LMT")

	inOutTest(c, `{{ (time.Unix 1234567890).UTC.Format "2006-01-02 15 -0700" }}`,
		"2009-02-13 23 +0000")

	inOutTest(c, `{{ (time.Unix "1234567890").UTC.Format "2006-01-02 15 -0700" }}`,
		"2009-02-13 23 +0000")
}
