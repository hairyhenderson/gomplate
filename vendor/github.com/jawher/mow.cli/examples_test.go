package cli

import (
	"fmt"
	"os"
	"time"
)

func Example_greet() {
	app := App("greet", "Greet")
	app.Spec = "[NAME]"
	name := app.String(StringArg{Name: "NAME", Value: "stranger", Desc: "Your name", EnvVar: "USER"})
	app.Action = func() {
		fmt.Printf("Hello %s\n", *name)
	}
	app.Run(os.Args)
}

func Example_cp() {
	cp := App("cp", "Copy files around")
	cp.Spec = "[-R [-H | -L | -P]] [-fi | -n] SRC... DST"

	var (
		recursive = cp.Bool(BoolOpt{
			Name:  "R",
			Value: false,
			Desc:  "copy src files recursively",
		})

		followSymbolicCL   = cp.Bool(BoolOpt{Name: "H", Value: false, Desc: "If the -R option is specified, symbolic links on the command line are followed.  (Symbolic links encountered in the tree traversal are not followed.)"})
		followSymbolicTree = cp.Bool(BoolOpt{Name: "L", Value: false, Desc: "If the -R option is specified, all symbolic links are followed."})
		followSymbolicNo   = cp.Bool(BoolOpt{Name: "P", Value: true, Desc: "If the -R option is specified, no symbolic links are followed.  This is the default."})

		force       = cp.Bool(BoolOpt{Name: "f", Value: false, Desc: "If the destination file cannot be opened, remove it and create a new file, without prompting for confirmation regardless of its permissions.  (The -f option overrides any previous -n option.)"})
		interactive = cp.Bool(BoolOpt{Name: "i", Value: false, Desc: "Cause cp to write a prompt to the standard error output before copying a file that would overwrite an existing file.  If the response from the standard input begins with the character `y' or `Y', the file copy is attempted.  (The -i option overrides any previous -n option.)"})
		noOverwrite = cp.Bool(BoolOpt{Name: "f", Value: false, Desc: "Do not overwrite an existing file.  (The -n option overrides any previous -f or -i options.)"})
	)

	var (
		src = cp.Strings(StringsArg{
			Name: "SRC",
			Desc: "The source files to copy",
		})
		dst = cp.Strings(StringsArg{Name: "DST", Value: nil, Desc: "The destination directory"})
	)

	cp.Action = func() {
		fmt.Printf(`copy:
	SRC: %v
	DST: %v
	recursive: %v
	follow links (CL, Tree, No): %v %v %v
	force: %v
	interactive: %v
	no overwrite: %v`,
			*src, *dst, *recursive,
			*followSymbolicCL, *followSymbolicTree, *followSymbolicNo,
			*force,
			*interactive,
			*noOverwrite)
	}

	cp.Run(os.Args)
}

func Example_docker() {
	docker := App("docker", "A self-sufficient runtime for linux containers")

	docker.Command("run", "Run a command in a new container", func(cmd *Cmd) {
		cmd.Spec = "[-d|--rm] IMAGE [COMMAND [ARG...]]"

		var (
			detached = cmd.Bool(BoolOpt{Name: "d detach", Value: false, Desc: "Detached mode: run the container in the background and print the new container ID"})
			rm       = cmd.Bool(BoolOpt{Name: "rm", Value: false, Desc: "Automatically remove the container when it exits (incompatible with -d)"})
			memory   = cmd.String(StringOpt{Name: "m memory", Value: "", Desc: "Memory limit (format: <number><optional unit>, where unit = b, k, m or g)"})
		)

		var (
			image   = cmd.String(StringArg{Name: "IMAGE", Value: "", Desc: ""})
			command = cmd.String(StringArg{Name: "COMMAND", Value: "", Desc: "The command to run"})
			args    = cmd.Strings(StringsArg{Name: "ARG", Value: nil, Desc: "The command arguments"})
		)

		cmd.Action = func() {
			var how string
			switch {
			case *detached:
				how = "detached"
			case *rm:
				how = "rm after"
			default:
				how = "--"
			}
			fmt.Printf("Run image %s, command %s, args %v, how? %v, mem %s", *image, *command, *args, how, *memory)
		}
	})

	docker.Command("pull", "Pull an image or a repository from the registry", func(cmd *Cmd) {
		cmd.Spec = "[-a] NAME"

		all := cmd.Bool(BoolOpt{Name: "a all-tags", Value: false, Desc: "Download all tagged images in the repository"})

		name := cmd.String(StringArg{Name: "NAME", Value: "", Desc: "Image name (optionally NAME:TAG)"})

		cmd.Action = func() {
			if *all {
				fmt.Printf("Download all tags for image %s", *name)
				return
			}
			fmt.Printf("Download image %s", *name)
		}
	})

	docker.Run(os.Args)
}

func Example_beforeAfter() {
	app := App("app", "App")
	bench := app.BoolOpt("b bench", false, "Measure execution time")

	var t0 time.Time

	app.Before = func() {
		if *bench {
			t0 = time.Now()
		}
	}

	app.After = func() {
		if *bench {
			d := time.Since(t0)
			fmt.Printf("Command execution took: %vs", d.Seconds())
		}
	}

	app.Command("cmd1", "first command", func(cmd *Cmd) {
		cmd.Action = func() {
			fmt.Print("Running command 1")
		}
	})

	app.Command("cmd2", "second command", func(cmd *Cmd) {
		cmd.Action = func() {
			fmt.Print("Running command 2")
		}
	})

	app.Run(os.Args)
}
