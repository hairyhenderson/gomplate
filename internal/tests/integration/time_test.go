package integration

import (
	"testing"
	"time"
	_ "time/tzdata"
)

func TestTime(t *testing.T) {
	f := `Mon Jan 02 15:04:05 MST 2006`
	i := `Fri Feb 13 23:31:30 UTC 2009`
	inOutTest(t, `{{ (time.Parse "`+f+`" "`+i+`").Format "2006-01-02 15 -0700" }}`,
		"2009-02-13 23 +0000")

	if !isWindows {
		o, e, err := cmd(t, "-i", `{{ time.ZoneName }}`).withEnv("TZ", "UTC").run()
		assertSuccess(t, o, e, err, "UTC")

		o, e, err = cmd(t, "-i", `{{ time.ZoneOffset }}`).withEnv("TZ", "UTC").run()
		assertSuccess(t, o, e, err, "0")

		o, e, err = cmd(t, "-i",
			`{{ (time.ParseLocal time.Kitchen "6:00AM").Format "15:04 MST" }}`).
			withEnv("TZ", "Africa/Luanda").run()
		assertSuccess(t, o, e, err, "06:00 LMT")
	}

	zname, _ := time.Now().Zone()
	inOutTest(t, `{{ time.ZoneName }}`, zname)

	inOutTest(t, `{{ (time.Now).Format "2006-01-02 15 -0700" }}`,
		time.Now().Format("2006-01-02 15 -0700"))

	inOutTest(t, `{{ (time.ParseInLocation time.Kitchen "Africa/Luanda" "6:00AM").Format "15:04 MST" }}`,
		"06:00 LMT")

	inOutTest(t, `{{ (time.Unix 1234567890).UTC.Format "2006-01-02 15 -0700" }}`,
		"2009-02-13 23 +0000")

	inOutTest(t, `{{ (time.Unix "1234567890").UTC.Format "2006-01-02 15 -0700" }}`,
		"2009-02-13 23 +0000")
}
