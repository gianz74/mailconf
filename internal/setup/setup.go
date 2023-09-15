package setup

import (
	"embed"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gianz74/mailconf"
	"github.com/gianz74/mailconf/internal/base"
	"github.com/gianz74/mailconf/internal/config"
	"github.com/gianz74/mailconf/internal/io"
	"github.com/gianz74/mailconf/internal/myterm"
	"github.com/gianz74/mailconf/internal/options"
	"github.com/gianz74/mailconf/internal/os"
)

var CmdSetup = &base.Command{
	UsageLine: "setup [-dry-run -v]",
	Short:     "setup configures email accounts",
	Long: `
Setup checks if the prerequisites for configuring email accounts are
met.

It also provides some guidance to satisfy the requirements.  If the
requirements are met, it prepares the system copying some scripts to
the locations provided by the user.

It optionally allows the user to specify the email profiles to be
configured.

The -dry-run option allows the user to preview the changes without
actually making any to the system.

The -v option increases verbosity, printing the content of the files
that are to be written.`,
}
var (
	dryrun  bool
	verbose bool

	//go:embed onnewmail.sh
	//go:embed syncmail.sh
	f               embed.FS
	ErrExists       = errors.New("Config file exists.")
	ErrRequirements = errors.New("Requirements not met.")
	ErrNoTerm       = errors.New("Not in a terminal.")
)

func init() {
	CmdSetup.Run = runSetup
	CmdSetup.Flag.BoolVar(&dryrun, "dry-run", false, "Show changes without making any.")
	CmdSetup.Flag.BoolVar(&verbose, "v", false, "Show content of files to be written.")
}

func runSetup(cmd *base.Command, args []string) error {
	options.Set(options.OptDryrun(dryrun), options.OptVerbose(verbose))
	cfg := config.Read()
	if cfg != nil {
		return ErrExists
	}

	cfg = config.NewConfig()

	t := myterm.New()
	emacsdir, err := t.ReadLine("enter emacs config directory: ")
	if err != nil {
		return err
	}
	cfg.EmacsCfgDir = expandUser(emacsdir)
	bindir, err := t.ReadLine("enter user's bin directory: ")
	if err != nil {
		return err
	}

	cfg.BinDir = expandUser(bindir)
	if !checkRequirements(cfg.BinDir) {
		return ErrRequirements
	}

	data, _ := f.ReadFile("onnewmail.sh")
	io.Write(filepath.Join(cfg.BinDir, "onnewmail.sh"), data, 0750)

	data, _ = f.ReadFile("syncmail.sh")
	io.Write(filepath.Join(cfg.BinDir, "syncmail.sh"), data, 0750)

	ans, err := t.ReadLine("do you want to create an email profile? [y/n]: ")
	if err != nil {
		return err
	}

	if len(ans) == 0 {
		cfg.Save()
		return nil
	}
	if !(ans[0] == 'y' || ans[0] == 'Y') {
		cfg.Save()
		return nil
	}

	for {
		profile, err := t.ReadLine("Profile name: ")
		if err != nil {
			return err
		}

		err = mailconf.AddProfile(profile, cfg)
		if err != nil {
			return err
		}

		ans, err := t.ReadLine("do you want to create another profile? [y/n]: ")
		if err != nil {
			return err
		}

		if len(ans) == 0 {
			break
		}
		if !(ans[0] == 'y' || ans[0] == 'Y') {
			break
		}
	}
	err = cfg.Save()
	if err != nil {
		return err
	}
	return nil
}

var checkRequirements = _checkRequirements

func _checkRequirements(bindir string) bool {
	if os.System == "linux" {
		_, err := exec.LookPath("secret-tool")
		if err != nil {
			fmt.Fprintf(os.Stderr, "secret-tool not found in PATH.\n")
			fmt.Fprintf(os.Stderr, "On Debian: install it with: apt install libsecret-tools\n")
			return false
		}
	}
	_, err := exec.LookPath("mbsync")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mbsync not found in PATH.\n")
		fmt.Fprintf(os.Stderr, "On Debian: install it with: apt install isync\n")
		fmt.Fprintf(os.Stderr, "On MacOSX: install it with: brew install isync\n")
		return false
	}
	_, err = exec.LookPath("mu")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mu not found in PATH.\n")
		fmt.Fprintf(os.Stderr, "On Debian: install it with: apt install maildir-utils\n")
		fmt.Fprintf(os.Stderr, "On MacOSX: install it with: brew install mu\n")
		return false
	}
	_, err = exec.LookPath(filepath.Join(bindir, "goimapnotify"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "goimapnotify not found in %s.\n", bindir)
		fmt.Fprintf(os.Stderr, "please install it with \"GOBIN=%s go install gitlab.com/shackra/goimapnotify@latest\"\n", bindir)
		return false
	}
	return true
}

func expandUser(path string) string {
	home, _ := os.UserHomeDir()

	if path == "~" {
		return home
	} else if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}
