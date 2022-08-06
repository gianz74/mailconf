package add

import (
	"bufio"
	"errors"
	"fmt"

	"github.com/gianz74/mailconf"
	"github.com/gianz74/mailconf/internal/base"
	"github.com/gianz74/mailconf/internal/config"
	"github.com/gianz74/mailconf/internal/os"
	"golang.org/x/term"
)

var CmdAdd = &base.Command{
	UsageLine: "add [-dry-run -v]",
	Short:     "add creates a new profile",
	Long: `

Add creates a new profile, asking the user to provide the required
information.

The -dry-run option allows the user to preview the changes without
actually making any to the system.

The -v option increases verbosity, printing the content of the files
that are to be written.`,
}

var (
	dryrun      bool
	verbose     bool
	ErrNoConfig = errors.New("Missing config file.")
	ErrNoTerm   = errors.New("Not in a terminal.")
)

func init() {
	CmdAdd.Run = runAdd
	CmdAdd.Flag.BoolVar(&dryrun, "dry-run", false, "Show changes without making any.")
	CmdAdd.Flag.BoolVar(&verbose, "v", false, "Show content of files to be written.")
}

func runAdd(cmd *base.Command, args []string) error {
	cfg := config.Read()
	if cfg == nil {
		fmt.Fprintf(os.Stderr, "missing config: run \"mailconf setup\" first.")
		return ErrNoConfig
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Profile name: ")
	scanner.Scan()
	profile := scanner.Text()

	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Fprintf(os.Stderr, "not in a terminal\n")
		return ErrNoTerm
	}
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))

	err = mailconf.AddProfile(profile, cfg)
	if err != nil {
		term.Restore(int(os.Stdin.Fd()), oldState)
		fmt.Fprintf(os.Stderr, "Cannot create profile: %v\n", err)
		return err
	}
	cfg.Save()
	term.Restore(int(os.Stdin.Fd()), oldState)
	return nil
}
