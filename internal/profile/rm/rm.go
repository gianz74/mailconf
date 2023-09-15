package rm

import (
	"errors"
	"fmt"
	"os"

	"github.com/gianz74/mailconf"
	"github.com/gianz74/mailconf/internal/base"
	"github.com/gianz74/mailconf/internal/config"
	"github.com/gianz74/mailconf/internal/myterm"
	"github.com/gianz74/mailconf/internal/options"
)

var CmdRm = &base.Command{
	UsageLine: "rm [-dry-run -v]",
	Short:     "rm deletes a profile",
	Long: `

Rm deletes a profile, asking the user to provide the required
information.

The -dry-run option allows the user to preview the changes without
actually making any to the system.

The -v option increases verbosity, printing the actions that are about
to be taken.`,
}

var (
	dryrun      bool
	verbose     bool
	ErrNoConfig = errors.New("Missing config file.")
)

func init() {
	CmdRm.Run = runRm
	CmdRm.Flag.BoolVar(&dryrun, "dry-run", false, "Show changes without making any.")
	CmdRm.Flag.BoolVar(&verbose, "v", false, "Print actions about to be taken.")
}

func runRm(cmd *base.Command, args []string) error {
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

	err = mailconf.RmProfile(profile, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot remove profile: %v\n", err)
		return err
	}
	cfg.Save()
	return nil
}
