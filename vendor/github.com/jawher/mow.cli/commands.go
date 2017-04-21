package cli

import (
	"bytes"
	"flag"
	"fmt"
	"strings"
	"text/tabwriter"
)

/*
Cmd represents a command (or sub command) in a CLI application. It should be constructed
by calling Command() on an app to create a top level command or by calling Command() on another
command to create a sub command
*/
type Cmd struct {
	// The code to execute when this command is matched
	Action func()
	// The code to execute before this command or any of its children is matched
	Before func()
	// The code to execute after this command or any of its children is matched
	After func()
	// The command options and arguments
	Spec string
	// The command long description to be shown when help is requested
	LongDesc string
	// The command error handling strategy
	ErrorHandling flag.ErrorHandling

	init    CmdInitializer
	name    string
	aliases []string
	desc    string

	commands   []*Cmd
	options    []*opt
	optionsIdx map[string]*opt
	args       []*arg
	argsIdx    map[string]*arg

	parents []string

	fsm *state
}

/*
BoolParam represents a Bool option or argument
*/
type BoolParam interface {
	value() bool
}

/*
StringParam represents a String option or argument
*/
type StringParam interface {
	value() string
}

/*
IntParam represents an Int option or argument
*/
type IntParam interface {
	value() int
}

/*
StringsParam represents a string slice option or argument
*/
type StringsParam interface {
	value() []string
}

/*
IntsParam represents an int slice option or argument
*/
type IntsParam interface {
	value() []int
}

/*
VarParam represents an custom option or argument where the type and format are controlled by the developer
*/
type VarParam interface {
	value() flag.Value
}

/*
CmdInitializer is a function that configures a command by adding options, arguments, a spec, sub commands and the code
to execute when the command is called
*/
type CmdInitializer func(*Cmd)

/*
Command adds a new (sub) command to c where name is the command name (what you type in the console),
description is what would be shown in the help messages, e.g.:

	Usage: git [OPTIONS] COMMAND [arg...]

	Commands:
	  $name	$desc

the last argument, init, is a function that will be called by mow.cli to further configure the created
(sub) command, e.g. to add options, arguments and the code to execute
*/
func (c *Cmd) Command(name, desc string, init CmdInitializer) {
	aliases := strings.Split(name, " ")
	c.commands = append(c.commands, &Cmd{
		ErrorHandling: c.ErrorHandling,
		name:          aliases[0],
		aliases:       aliases,
		desc:          desc,
		init:          init,
		commands:      []*Cmd{},
		options:       []*opt{},
		optionsIdx:    map[string]*opt{},
		args:          []*arg{},
		argsIdx:       map[string]*arg{},
	})
}

/*
Bool can be used to add a bool option or argument to a command.
It accepts either a BoolOpt or a BoolArg struct.

The result should be stored in a variable (a pointer to a bool) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) Bool(p BoolParam) *bool {
	into := new(bool)
	value := newBoolValue(into, p.value())

	switch x := p.(type) {
	case BoolOpt:
		c.mkOpt(opt{name: x.Name, desc: x.Desc, envVar: x.EnvVar, hideValue: x.HideValue, value: value, valueSetByUser: x.SetByUser})
	case BoolArg:
		c.mkArg(arg{name: x.Name, desc: x.Desc, envVar: x.EnvVar, hideValue: x.HideValue, value: value, valueSetByUser: x.SetByUser})
	default:
		panic(fmt.Sprintf("Unhandled param %v", p))
	}

	return into
}

/*
String can be used to add a string option or argument to a command.
It accepts either a StringOpt or a StringArg struct.

The result should be stored in a variable (a pointer to a string) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) String(p StringParam) *string {
	into := new(string)
	value := newStringValue(into, p.value())

	switch x := p.(type) {
	case StringOpt:
		c.mkOpt(opt{name: x.Name, desc: x.Desc, envVar: x.EnvVar, hideValue: x.HideValue, value: value, valueSetByUser: x.SetByUser})
	case StringArg:
		c.mkArg(arg{name: x.Name, desc: x.Desc, envVar: x.EnvVar, hideValue: x.HideValue, value: value, valueSetByUser: x.SetByUser})
	default:
		panic(fmt.Sprintf("Unhandled param %v", p))
	}

	return into
}

/*
Int can be used to add an int option or argument to a command.
It accepts either a IntOpt or a IntArg struct.

The result should be stored in a variable (a pointer to an int) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) Int(p IntParam) *int {
	into := new(int)
	value := newIntValue(into, p.value())

	switch x := p.(type) {
	case IntOpt:
		c.mkOpt(opt{name: x.Name, desc: x.Desc, envVar: x.EnvVar, hideValue: x.HideValue, value: value, valueSetByUser: x.SetByUser})
	case IntArg:
		c.mkArg(arg{name: x.Name, desc: x.Desc, envVar: x.EnvVar, hideValue: x.HideValue, value: value, valueSetByUser: x.SetByUser})
	default:
		panic(fmt.Sprintf("Unhandled param %v", p))
	}

	return into
}

/*
Strings can be used to add a string slice option or argument to a command.
It accepts either a StringsOpt or a StringsArg struct.

The result should be stored in a variable (a pointer to a string slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) Strings(p StringsParam) *[]string {
	into := new([]string)
	value := newStringsValue(into, p.value())

	switch x := p.(type) {
	case StringsOpt:
		c.mkOpt(opt{name: x.Name, desc: x.Desc, envVar: x.EnvVar, hideValue: x.HideValue, value: value, valueSetByUser: x.SetByUser})
	case StringsArg:
		c.mkArg(arg{name: x.Name, desc: x.Desc, envVar: x.EnvVar, hideValue: x.HideValue, value: value, valueSetByUser: x.SetByUser})
	default:
		panic(fmt.Sprintf("Unhandled param %v", p))
	}

	return into
}

/*
Ints can be used to add an int slice option or argument to a command.
It accepts either a IntsOpt or a IntsArg struct.

The result should be stored in a variable (a pointer to an int slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) Ints(p IntsParam) *[]int {
	into := new([]int)
	value := newIntsValue(into, p.value())

	switch x := p.(type) {
	case IntsOpt:
		c.mkOpt(opt{name: x.Name, desc: x.Desc, envVar: x.EnvVar, hideValue: x.HideValue, value: value, valueSetByUser: x.SetByUser})
	case IntsArg:
		c.mkArg(arg{name: x.Name, desc: x.Desc, envVar: x.EnvVar, hideValue: x.HideValue, value: value, valueSetByUser: x.SetByUser})
	default:
		panic(fmt.Sprintf("Unhandled param %v", p))
	}

	return into
}

/*
Var can be used to add a custom option or argument to a command.
It accepts either a VarOpt or a VarArg struct.

As opposed to the other built-in types, this function does not return a pointer the the value.
Instead, the VarOpt or VarOptArg structs hold the said value.
*/
func (c *Cmd) Var(p VarParam) {
	switch x := p.(type) {
	case VarOpt:
		c.mkOpt(opt{name: x.Name, desc: x.Desc, envVar: x.EnvVar, hideValue: x.HideValue, value: p.value(), valueSetByUser: x.SetByUser})
	case VarArg:
		c.mkArg(arg{name: x.Name, desc: x.Desc, envVar: x.EnvVar, hideValue: x.HideValue, value: p.value(), valueSetByUser: x.SetByUser})
	default:
		panic(fmt.Sprintf("Unhandled param %v", p))
	}
}

func (c *Cmd) doInit() error {
	if c.init != nil {
		c.init(c)
	}

	parents := append(c.parents, c.name)

	for _, sub := range c.commands {
		sub.parents = parents
	}

	if len(c.Spec) == 0 {
		if len(c.options) > 0 {
			c.Spec = "[OPTIONS] "
		}
		for _, arg := range c.args {
			c.Spec += arg.name + " "
		}
	}
	fsm, err := uParse(c)
	if err != nil {
		return err
	}
	c.fsm = fsm
	return nil
}

func (c *Cmd) onError(err error) {
	if err != nil {
		switch c.ErrorHandling {
		case flag.ExitOnError:
			exiter(2)
		case flag.PanicOnError:
			panic(err)
		}
	} else {
		if c.ErrorHandling == flag.ExitOnError {
			exiter(2)
		}
	}
}

/*
PrintHelp prints the command's help message.
In most cases the library users won't need to call this method, unless
a more complex validation is needed
*/
func (c *Cmd) PrintHelp() {
	c.printHelp(false)
}

/*
PrintLongHelp prints the command's help message using the command long description if specified.
In most cases the library users won't need to call this method, unless
a more complex validation is needed
*/
func (c *Cmd) PrintLongHelp() {
	c.printHelp(true)
}

func (c *Cmd) printHelp(longDesc bool) {
	full := append(c.parents, c.name)
	path := strings.Join(full, " ")
	fmt.Fprintf(stdErr, "\nUsage: %s", path)

	spec := strings.TrimSpace(c.Spec)
	if len(spec) > 0 {
		fmt.Fprintf(stdErr, " %s", spec)
	}

	if len(c.commands) > 0 {
		fmt.Fprint(stdErr, " COMMAND [arg...]")
	}
	fmt.Fprint(stdErr, "\n\n")

	desc := c.desc
	if longDesc && len(c.LongDesc) > 0 {
		desc = c.LongDesc
	}
	if len(desc) > 0 {
		fmt.Fprintf(stdErr, "%s\n", desc)
	}

	w := tabwriter.NewWriter(stdErr, 15, 1, 3, ' ', 0)

	if len(c.args) > 0 {
		fmt.Fprintf(stdErr, "\nArguments:\n")

		for _, arg := range c.args {
			desc := c.formatDescription(arg.desc, arg.envVar)
			value := c.formatArgValue(arg)

			fmt.Fprintf(w, "  %s%s\t%s\n", arg.name, value, desc)
		}
		w.Flush()
	}

	if len(c.options) > 0 {
		fmt.Fprintf(stdErr, "\nOptions:\n")

		for _, opt := range c.options {
			desc := c.formatDescription(opt.desc, opt.envVar)
			value := c.formatOptValue(opt)
			fmt.Fprintf(w, "  %s%s\t%s\n", strings.Join(opt.names, ", "), value, desc)
		}
		w.Flush()
	}

	if len(c.commands) > 0 {
		fmt.Fprintf(stdErr, "\nCommands:\n")

		for _, c := range c.commands {
			fmt.Fprintf(w, "  %s\t%s\n", strings.Join(c.aliases, ", "), c.desc)
		}
		w.Flush()
	}

	if len(c.commands) > 0 {
		fmt.Fprintf(stdErr, "\nRun '%s COMMAND --help' for more information on a command.\n", path)
	}
}

func (c *Cmd) formatArgValue(arg *arg) string {
	if arg.hideValue {
		return " "
	}
	return "=" + arg.value.String()
}

func (c *Cmd) formatOptValue(opt *opt) string {
	if opt.hideValue {
		return " "
	}
	return "=" + opt.value.String()
}

func (c *Cmd) formatDescription(desc, envVar string) string {
	var b bytes.Buffer
	b.WriteString(desc)
	if len(envVar) > 0 {
		b.WriteString(" (")
		sep := ""
		for _, envVal := range strings.Split(envVar, " ") {
			b.WriteString(fmt.Sprintf("%s$%s", sep, envVal))
			sep = " "
		}
		b.WriteString(")")
	}
	return strings.TrimSpace(b.String())
}

func (c *Cmd) parse(args []string, entry, inFlow, outFlow *step) error {
	if c.helpRequested(args) {
		c.PrintLongHelp()
		c.onError(nil)
		return nil
	}

	nargsLen := c.getOptsAndArgs(args)

	if err := c.fsm.parse(args[:nargsLen]); err != nil {
		fmt.Fprintf(stdErr, "Error: %s\n", err.Error())
		c.PrintHelp()
		c.onError(err)
		return err
	}

	newInFlow := &step{
		do:    c.Before,
		error: outFlow,
		desc:  fmt.Sprintf("%s.Before", c.name),
	}
	inFlow.success = newInFlow

	newOutFlow := &step{
		do:      c.After,
		success: outFlow,
		error:   outFlow,
		desc:    fmt.Sprintf("%s.After", c.name),
	}

	args = args[nargsLen:]
	if len(args) == 0 {
		if c.Action != nil {
			newInFlow.success = &step{
				do:      c.Action,
				success: newOutFlow,
				error:   newOutFlow,
				desc:    fmt.Sprintf("%s.Action", c.name),
			}

			entry.run(nil)
			return nil
		}
		c.PrintHelp()
		c.onError(nil)
		return nil
	}

	arg := args[0]
	for _, sub := range c.commands {
		if sub.isAlias(arg) {
			if err := sub.doInit(); err != nil {
				panic(err)
			}
			return sub.parse(args[1:], entry, newInFlow, newOutFlow)
		}
	}

	var err error
	switch {
	case strings.HasPrefix(arg, "-"):
		err = fmt.Errorf("Error: illegal option %s", arg)
		fmt.Fprintln(stdErr, err.Error())
	default:
		err = fmt.Errorf("Error: illegal input %s", arg)
		fmt.Fprintln(stdErr, err.Error())
	}
	c.PrintHelp()
	c.onError(err)
	return err

}

func (c *Cmd) isArgSet(args []string, searchArgs []string) bool {
	for _, arg := range args {
		for _, sub := range c.commands {
			if sub.isAlias(arg) {
				return false
			}
		}
		for _, searchArg := range searchArgs {
			if arg == searchArg {
				return true
			}
		}
	}
	return false
}

func (c *Cmd) helpRequested(args []string) bool {
	return c.isArgSet(args, []string{"-h", "--help"})
}

func (c *Cmd) getOptsAndArgs(args []string) int {
	consumed := 0

	for _, arg := range args {
		for _, sub := range c.commands {
			if sub.isAlias(arg) {
				return consumed
			}
		}
		consumed++
	}
	return consumed
}

func (c *Cmd) isAlias(arg string) bool {
	for _, alias := range c.aliases {
		if arg == alias {
			return true
		}
	}
	return false
}
