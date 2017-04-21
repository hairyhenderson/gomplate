package cli

import (
	"fmt"
	"testing"
	"time"

	"os"

	"github.com/stretchr/testify/require"
)

// Counter
type Counter int

func (d *Counter) Set(v string) error {
	*d++
	return nil
}

func (d *Counter) String() string {
	return fmt.Sprintf("%d", *d)
}

func (d *Counter) IsBoolFlag() bool {
	return true
}

// Duration
type Duration time.Duration

func (d *Duration) Set(v string) error {
	parsed, err := time.ParseDuration(v)
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

func (d *Duration) String() string {
	duration := time.Duration(*d)
	return duration.String()
}

type Percent []float64

func parsePercent(v string) (float64, error) {
	var d int
	_, err := fmt.Sscanf(v, "%d%%", &d)
	if err != nil {
		return 0, err
	}
	return float64(d) / 100, nil
}

func (p *Percent) Set(v string) error {
	f, err := parsePercent(v)
	if err != nil {
		return err
	}
	*p = append(*p, f)
	return nil
}

func (p *Percent) Clear() {
	*p = nil
}

func (p *Percent) String() string {
	res := "["
	for idx, p := range *p {
		if idx > 0 {
			res += ", "
		}
		res += fmt.Sprintf("%.0f%%", p*100)
	}
	return res + "]"
}

func TestVar(t *testing.T) {
	value := Counter(0)
	duration := Duration(0)
	percents := Percent{}

	app := App("var", "")
	app.Spec = "-v... DURATION PERCENT..."

	app.VarOpt("v", &value, "")
	app.VarArg("DURATION", &duration, "")
	app.VarArg("PERCENT", &percents, "")

	ex := false
	app.Action = func() {
		ex = true
	}
	app.Run([]string{"cp", "-vvv", "1h", "10%", "5%"})

	require.Equal(t, Counter(3), value)
	require.Equal(t, Duration(1*time.Hour), duration)
	require.Equal(t, Percent([]float64{0.1, 0.05}), percents)

	require.True(t, ex, "Exec wasn't called")
}

func TestVarFromEnv(t *testing.T) {
	os.Setenv("MOWCLI_DURATION", "1h2m3s")
	os.Setenv("MOWCLI_ARG_PERCENTS", "25%, 1%")
	os.Setenv("MOWCLI_OPT_PERCENTS", "90%, 42%")

	duration := Duration(0)
	argPercents := Percent{}
	optPercents := Percent{}

	app := App("var", "")
	app.Spec = "-p... DURATION PERCENT..."

	app.Var(VarArg{
		Name:   "DURATION",
		Value:  &duration,
		EnvVar: "MOWCLI_DURATION",
	})
	app.Var(VarArg{
		Name:   "PERCENT",
		Value:  &argPercents,
		EnvVar: "MOWCLI_ARG_PERCENTS",
	})
	app.Var(VarOpt{
		Name:   "p",
		Value:  &optPercents,
		EnvVar: "MOWCLI_OPT_PERCENTS",
	})

	require.Equal(t, Duration(1*time.Hour+2*time.Minute+3*time.Second), duration)

	require.Equal(t, Percent([]float64{0.25, 0.01}), argPercents)
	require.Equal(t, Percent([]float64{0.9, 0.42}), optPercents)
}

func TestVarOverrideEnv(t *testing.T) {
	os.Setenv("MOWCLI_PERCENTS", "25%, 1%")

	percents := Percent{}

	app := App("var", "")
	app.Spec = "PERCENT..."
	app.Var(VarArg{
		Name:   "PERCENT",
		Value:  &percents,
		EnvVar: "MOWCLI_PERCENTS",
	})

	ex := false
	app.Action = func() {
		ex = true
		require.Equal(t, Percent([]float64{0, 0.99}), percents)
	}

	app.Run([]string{"var", "0%", "99%"})

	require.True(t, ex, "Action should have been called")
}
