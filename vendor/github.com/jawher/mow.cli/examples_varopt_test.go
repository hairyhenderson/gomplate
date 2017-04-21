package cli_test

import (
	"fmt"
	"time"

	cli "github.com/jawher/mow.cli"
)

// Declare your type
type Duration time.Duration

// Make it implement flag.Value
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

func ExampleVarArg() {

	app := cli.App("var", "Var arg example")

	// Declare a variable of your type
	duration := Duration(0)
	// Call one of the Var methods (arg, opt, ...) to declare your custom type
	app.VarArg("DURATION", &duration, "")

	app.Action = func() {
		// The variable will be populated after the app is ran
		fmt.Print(time.Duration(duration))
	}

	app.Run([]string{"cp", "1h31m42s"})
	// Output: 1h31m42s
}
