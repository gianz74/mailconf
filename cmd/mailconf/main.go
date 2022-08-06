package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gianz74/mailconf/internal/base"
	"github.com/gianz74/mailconf/internal/help"
	"github.com/gianz74/mailconf/internal/os"
	"github.com/gianz74/mailconf/internal/profile"
	"github.com/gianz74/mailconf/internal/setup"
)

func init() {
	base.Commands = []*base.Command{
		setup.CmdSetup,
		profile.CmdProfile,
	}
	base.Usage = mainUsage
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Usage = base.Usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		base.Usage()
	}

	if args[0] == "help" {
		help.Help(base.Commands, args[1:], []*base.Command{})
		return
	}

	for _, cmd := range base.Commands {
		cmd.Flag.Usage = cmd.Usage
		if cmd.Name() == args[0] {
			cmd.Flag.Parse(args[1:])
			args = cmd.Flag.Args()
			cmd.Run(cmd, args)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "mailconf: unknown subcommand %q\nRun 'mailconf help' for usage.\n", args[0])
	os.Exit(2)
}

func mainUsage() {
	help.PrintUsage()
	os.Exit(2)
}
