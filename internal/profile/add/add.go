package add

import (
	"errors"
	"fmt"

	"github.com/gianz74/mailconf"
	"github.com/gianz74/mailconf/internal/base"
	"github.com/gianz74/mailconf/internal/config"
	"github.com/gianz74/mailconf/internal/myterm"
	"github.com/gianz74/mailconf/internal/options"
	"github.com/gianz74/mailconf/internal/os"
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
)

func init() {
	CmdAdd.Run = runAdd
	CmdAdd.Flag.BoolVar(&dryrun, "dry-run", false, "Show changes without making any.")
	CmdAdd.Flag.BoolVar(&verbose, "v", false, "Show content of files to be written.")
}

func runAdd(cmd *base.Command, args []string) error {
	options.Set(options.OptDryrun(dryrun), options.OptVerbose(verbose))
	cfg := config.Read()
	if cfg == nil {
		fmt.Fprintf(os.Stderr, "missing config: run \"mailconf setup\" first.")
		return ErrNoConfig
	}

	t := myterm.New()
	profile, err := t.ReadLine("Profile name: ")
	if err != nil {
		return err
	}

	err = mailconf.AddProfile(profile, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create profile: %v\n", err)
		return err
	}
	cfg.Save()
	return nil
}
