package integration

import (
	"os"
	"time"
	_ "time/tzdata"

	. "gopkg.in/check.v1"
)

type TimeSuite struct{}

var _ = Suite(&TimeSuite{})

// convenience function to run a test in a specific timezone
func runWithTZ(c *C, tz string, f func(c *C)) {
	origTZ := os.Getenv("TZ")
	defer func() { os.Setenv("TZ", origTZ) }()
	os.Setenv("TZ", tz)

	f(c)
}

func (s *TimeSuite) TestTime(c *C) {
	f := `Mon Jan 02 15:04:05 MST 2006`
	i := `Fri Feb 13 23:31:30 UTC 2009`
	inOutTest(c, `{{ (time.Parse "`+f+`" "`+i+`").Format "2006-01-02 15 -0700" }}`,
		"2009-02-13 23 +0000")

	if !isWindows {

		runWithTZ(c, "UTC", func(c *C) {
			inOutTest(c, `{{ time.ZoneName }}`, "UTC")
			inOutTest(c, `{{ time.ZoneOffset }}`, "0")
		})

		runWithTZ(c, "Africa/Luanda", func(c *C) {
			inOutTest(c, `{{ (time.ParseLocal time.Kitchen "6:00AM").Format "15:04 MST" }}`,
				"06:00 LMT")
		})
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
