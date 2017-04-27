package cli_test

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
)

// Declare your type
type Counter int

// Make it implement flag.Value
func (c *Counter) Set(v string) error {
	*c++
	return nil
}

func (c *Counter) String() string {
	return fmt.Sprintf("%d", *c)
}

// Make it a bool option
func (c *Counter) IsBoolFlag() bool {
	return true
}

func ExampleVarOpt() {

	app := cli.App("var", "Var opt example")

	// Declare a variable of your type
	verbosity := Counter(0)
	// Call one of the Var methods (arg, opt, ...) to declare your custom type
	app.VarOpt("v", &verbosity, "verbosity level")

	app.Action = func() {
		// The variable will be populated after the app is ran
		fmt.Print(verbosity)
	}

	app.Run([]string{"app", "-vvvvv"})
	// Output: 5
}
