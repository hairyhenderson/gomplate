package cli

import (
	"flag"

	"github.com/stretchr/testify/require"

	"os"
	"testing"
)

func TestTheCpCase(t *testing.T) {
	app := App("cp", "")
	app.Spec = "SRC... DST"

	src := app.Strings(StringsArg{Name: "SRC", Value: nil, Desc: ""})
	dst := app.String(StringArg{Name: "DST", Value: "", Desc: ""})

	ex := false
	app.Action = func() {
		ex = true
	}
	app.Run([]string{"cp", "x", "y", "z"})

	require.Equal(t, []string{"x", "y"}, *src)
	require.Equal(t, "z", *dst)

	require.True(t, ex, "Exec wasn't called")
}

func TestImplicitSpec(t *testing.T) {
	app := App("test", "")
	x := app.Bool(BoolOpt{Name: "x", Value: false, Desc: ""})
	y := app.String(StringOpt{Name: "y", Value: "", Desc: ""})
	called := false
	app.Action = func() {
		called = true
	}
	app.ErrorHandling = flag.ContinueOnError

	err := app.Run([]string{"test", "-x", "-y", "hello"})

	require.Nil(t, err)
	require.True(t, *x)
	require.Equal(t, "hello", *y)

	require.True(t, called, "Exec wasn't called")
}

func TestHelpShortcut(t *testing.T) {
	defer suppressOutput()()

	exitCalled := false
	defer exitShouldBeCalledWith(t, 2, &exitCalled)()

	app := App("x", "")
	app.Spec = "Y"

	app.String(StringArg{Name: "Y", Value: "", Desc: ""})

	actionCalled := false
	app.Action = func() {
		actionCalled = true
	}
	app.Run([]string{"x", "y", "-h", "z"})

	require.False(t, actionCalled, "action should not have been called")
	require.True(t, exitCalled, "exit should have been called")
}

func TestHelpMessage(t *testing.T) {
	var out, err string
	defer captureAndRestoreOutput(&out, &err)()

	exitCalled := false
	defer exitShouldBeCalledWith(t, 2, &exitCalled)()

	app := App("app", "App Desc")
	app.Spec = "[-o] ARG"

	app.String(StringOpt{Name: "o opt", Value: "", Desc: "Option"})
	app.String(StringArg{Name: "ARG", Value: "", Desc: "Argument"})

	app.Action = func() {}
	app.Run([]string{"app", "-h"})

	help := `
Usage: app [-o] ARG

App Desc

Arguments:
  ARG=""       Argument

Options:
  -o, --opt=""   Option
`

	require.Equal(t, help, err)
}

func TestLongHelpMessage(t *testing.T) {
	var out, err string
	defer captureAndRestoreOutput(&out, &err)()

	exitCalled := false
	defer exitShouldBeCalledWith(t, 2, &exitCalled)()

	app := App("app", "App Desc")
	app.LongDesc = "Longer App Desc"
	app.Spec = "[-o] ARG"

	app.String(StringOpt{Name: "o opt", Value: "", Desc: "Option"})
	app.String(StringArg{Name: "ARG", Value: "", Desc: "Argument"})

	app.Action = func() {}
	app.Run([]string{"app", "-h"})

	help := `
Usage: app [-o] ARG

Longer App Desc

Arguments:
  ARG=""       Argument

Options:
  -o, --opt=""   Option
`

	require.Equal(t, help, err)
}

func TestVersionShortcut(t *testing.T) {
	defer suppressOutput()()
	exitCalled := false
	defer exitShouldBeCalledWith(t, 0, &exitCalled)()

	app := App("cp", "")
	app.Version("v version", "cp 1.2.3")

	actionCalled := false
	app.Action = func() {
		actionCalled = true
	}

	app.Run([]string{"cp", "--version"})

	require.False(t, actionCalled, "action should not have been called")
	require.True(t, exitCalled, "exit should have been called")
}

func TestSubCommands(t *testing.T) {
	app := App("say", "")

	hi, bye := false, false

	app.Command("hi", "", func(cmd *Cmd) {
		cmd.Action = func() {
			hi = true
		}
	})

	app.Command("byte", "", func(cmd *Cmd) {
		cmd.Action = func() {
			bye = true
		}
	})

	app.Run([]string{"say", "hi"})
	require.True(t, hi, "hi should have been called")
	require.False(t, bye, "byte should NOT have been called")
}

func TestContinueOnError(t *testing.T) {
	defer exitShouldNotCalled(t)()
	defer suppressOutput()()

	app := App("say", "")
	app.String(StringOpt{Name: "f", Value: "", Desc: ""})
	app.Spec = "-f"
	app.ErrorHandling = flag.ContinueOnError
	called := false
	app.Action = func() {
		called = true
	}

	err := app.Run([]string{"say"})
	require.NotNil(t, err)
	require.False(t, called, "Exec should NOT have been called")
}

func TestExitOnError(t *testing.T) {
	defer suppressOutput()()

	exitCalled := false
	defer exitShouldBeCalledWith(t, 2, &exitCalled)()

	app := App("x", "")
	app.Spec = "Y"

	app.String(StringArg{Name: "Y", Value: "", Desc: ""})
	app.Run([]string{"x", "y", "z"})
	require.True(t, exitCalled, "exit should have been called")
}

func TestPanicOnError(t *testing.T) {
	defer suppressOutput()()

	app := App("say", "")
	app.String(StringOpt{Name: "f", Value: "", Desc: ""})
	app.Spec = "-f"
	app.ErrorHandling = flag.PanicOnError
	called := false
	app.Action = func() {
		called = true
	}

	defer func() {
		if r := recover(); r != nil {
			require.False(t, called, "Exec should NOT have been called")
		} else {

		}
	}()
	app.Run([]string{"say"})
	t.Fatalf("wanted panic")
}

func TestOptSetByUser(t *testing.T) {
	cases := []struct {
		desc     string
		config   func(*Cli, *bool)
		args     []string
		expected bool
	}{
		// OPTS
		// String
		{
			desc: "String Opt, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.String(StringOpt{Name: "f", Value: "a", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "String Opt, not set by user, env value",
			config: func(c *Cli, s *bool) {
				os.Setenv("MOW_VALUE", "value")
				c.String(StringOpt{Name: "f", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "String Opt, set by user",
			config: func(c *Cli, s *bool) {
				c.String(StringOpt{Name: "f", Value: "a", SetByUser: s})
			},
			args:     []string{"test", "-f=hello"},
			expected: true,
		},

		// Bool
		{
			desc: "Bool Opt, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Bool(BoolOpt{Name: "f", Value: true, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Bool Opt, not set by user, env value",
			config: func(c *Cli, s *bool) {
				os.Setenv("MOW_VALUE", "true")
				c.Bool(BoolOpt{Name: "f", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Bool Opt, set by user",
			config: func(c *Cli, s *bool) {
				c.Bool(BoolOpt{Name: "f", SetByUser: s})
			},
			args:     []string{"test", "-f"},
			expected: true,
		},

		// Int
		{
			desc: "Int Opt, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Int(IntOpt{Name: "f", Value: 42, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Int Opt, not set by user, env value",
			config: func(c *Cli, s *bool) {
				os.Setenv("MOW_VALUE", "33")
				c.Int(IntOpt{Name: "f", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Int Opt, set by user",
			config: func(c *Cli, s *bool) {
				c.Int(IntOpt{Name: "f", SetByUser: s})
			},
			args:     []string{"test", "-f=666"},
			expected: true,
		},

		// Ints
		{
			desc: "Ints Opt, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Ints(IntsOpt{Name: "f", Value: []int{42}, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Ints Opt, not set by user, env value",
			config: func(c *Cli, s *bool) {
				os.Setenv("MOW_VALUE", "11,22,33")
				c.Ints(IntsOpt{Name: "f", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Ints Opt, set by user",
			config: func(c *Cli, s *bool) {
				c.Ints(IntsOpt{Name: "f", SetByUser: s})
			},
			args:     []string{"test", "-f=666"},
			expected: true,
		},

		// Strings
		{
			desc: "Strings Opt, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Strings(StringsOpt{Name: "f", Value: []string{"aaa"}, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Strings Opt, not set by user, env value",
			config: func(c *Cli, s *bool) {
				os.Setenv("MOW_VALUE", "a,b,c")
				c.Strings(StringsOpt{Name: "f", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Strings Opt, set by user",
			config: func(c *Cli, s *bool) {
				c.Strings(StringsOpt{Name: "f", SetByUser: s})
			},
			args:     []string{"test", "-f=ccc"},
			expected: true,
		},
	}

	for _, cas := range cases {
		t.Log(cas.desc)

		setByUser := false
		app := App("test", "")

		cas.config(app, &setByUser)

		called := false
		app.Action = func() {
			called = true
		}

		app.Run(cas.args)

		require.True(t, called, "action should have been called")
		require.Equal(t, cas.expected, setByUser)
	}

}

func TestArgSetByUser(t *testing.T) {
	cases := []struct {
		desc     string
		config   func(*Cli, *bool)
		args     []string
		expected bool
	}{
		// ARGS
		// String
		{
			desc: "String Arg, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				c.String(StringArg{Name: "ARG", Value: "a", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "String Arg, not set by user, env value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				os.Setenv("MOW_VALUE", "value")
				c.String(StringArg{Name: "ARG", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "String Arg, set by user",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				c.String(StringArg{Name: "ARG", Value: "a", SetByUser: s})
			},
			args:     []string{"test", "aaa"},
			expected: true,
		},

		// Bool
		{
			desc: "Bool Arg, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				c.Bool(BoolArg{Name: "ARG", Value: true, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Bool Arg, not set by user, env value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				os.Setenv("MOW_VALUE", "true")
				c.Bool(BoolArg{Name: "ARG", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Bool Arg, set by user",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				c.Bool(BoolArg{Name: "ARG", SetByUser: s})
			},
			args:     []string{"test", "true"},
			expected: true,
		},

		// Int
		{
			desc: "Int Arg, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				c.Int(IntArg{Name: "ARG", Value: 42, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Int Arg, not set by user, env value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				os.Setenv("MOW_VALUE", "33")
				c.Int(IntArg{Name: "ARG", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Int Arg, set by user",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG]"
				c.Int(IntArg{Name: "ARG", SetByUser: s})
			},
			args:     []string{"test", "666"},
			expected: true,
		},

		// Ints
		{
			desc: "Ints Arg, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG...]"
				c.Ints(IntsArg{Name: "ARG", Value: []int{42}, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Ints Arg, not set by user, env value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG...]"
				os.Setenv("MOW_VALUE", "11,22,33")
				c.Ints(IntsArg{Name: "ARG", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Ints Arg, set by user",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG...]"
				c.Ints(IntsArg{Name: "ARG", SetByUser: s})
			},
			args:     []string{"test", "333", "666"},
			expected: true,
		},

		// Strings
		{
			desc: "Strings Arg, not set by user, default value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG...]"
				c.Strings(StringsArg{Name: "ARG", Value: []string{"aaa"}, SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Strings Arg, not set by user, env value",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG...]"
				os.Setenv("MOW_VALUE", "a,b,c")
				c.Strings(StringsArg{Name: "ARG", EnvVar: "MOW_VALUE", SetByUser: s})
			},
			args:     []string{"test"},
			expected: false,
		},
		{
			desc: "Strings Arg, set by user",
			config: func(c *Cli, s *bool) {
				c.Spec = "[ARG...]"
				c.Strings(StringsArg{Name: "ARG", SetByUser: s})
			},
			args:     []string{"test", "aaa", "ccc"},
			expected: true,
		},
	}

	for _, cas := range cases {
		t.Log(cas.desc)

		setByUser := false
		app := App("test", "")

		cas.config(app, &setByUser)

		called := false
		app.Action = func() {
			called = true
		}

		app.Run(cas.args)

		require.True(t, called, "action should have been called")
		require.Equal(t, cas.expected, setByUser)
	}

}

func TestCommandAliases(t *testing.T) {
	defer suppressOutput()()

	cases := []struct {
		args          []string
		errorExpected bool
	}{
		{
			args:          []string{"say", "hello"},
			errorExpected: false,
		},
		{
			args:          []string{"say", "hi"},
			errorExpected: false,
		},
		{
			args:          []string{"say", "hello hi"},
			errorExpected: true,
		},
		{
			args:          []string{"say", "hello", "hi"},
			errorExpected: true,
		},
	}

	for _, cas := range cases {
		app := App("say", "")
		app.ErrorHandling = flag.ContinueOnError

		called := false

		app.Command("hello hi", "", func(cmd *Cmd) {
			cmd.Action = func() {
				called = true
			}
		})

		err := app.Run(cas.args)

		if cas.errorExpected {
			require.Error(t, err, "Run() should have returned with an error")
			require.False(t, called, "action should not have been called")
		} else {
			require.NoError(t, err, "Run() should have returned without an error")
			require.True(t, called, "action should have been called")
		}
	}
}

func TestSubcommandAliases(t *testing.T) {
	cases := []struct {
		args []string
	}{
		{
			args: []string{"app", "foo", "bar", "baz"},
		},
		{
			args: []string{"app", "foo", "bar", "z"},
		},
		{
			args: []string{"app", "foo", "b", "baz"},
		},
		{
			args: []string{"app", "f", "bar", "baz"},
		},
		{
			args: []string{"app", "f", "b", "baz"},
		},
		{
			args: []string{"app", "f", "b", "z"},
		},
		{
			args: []string{"app", "foo", "b", "z"},
		},
		{
			args: []string{"app", "f", "bar", "z"},
		},
	}

	for _, cas := range cases {
		app := App("app", "")
		app.ErrorHandling = flag.ContinueOnError

		called := false

		app.Command("foo f", "", func(cmd *Cmd) {
			cmd.Command("bar b", "", func(cmd *Cmd) {
				cmd.Command("baz z", "", func(cmd *Cmd) {
					cmd.Action = func() {
						called = true
					}
				})
			})
		})

		err := app.Run(cas.args)

		require.NoError(t, err, "Run() should have returned without an error")
		require.True(t, called, "action should have been called")
	}
}
