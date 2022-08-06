package mailconf

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"path"
	"strconv"

	"github.com/gianz74/mailconf/internal/config"
	"github.com/gianz74/mailconf/internal/cred"
	"github.com/gianz74/mailconf/internal/io"
	"github.com/gianz74/mailconf/internal/myterm"
)

var (
	ErrProfileExists  = errors.New("Profile exists")
	ErrOsNotSupported = errors.New("OS not supported")
)

func AddProfile(profile string, cfg *config.Config) error {
	for _, p := range cfg.Profiles {
		if profile == p.Name {
			return ErrProfileExists
		}
	}
	p := &config.Profile{
		Name: profile,
	}

	t := myterm.New()
	var err error
	p.FullName, err = t.ReadLine("full user name: ")
	if err != nil {
		return err
	}

	p.Email, err = t.ReadLine("email address: ")
	if err != nil {
		return err
	}

	p.ImapHost, err = t.ReadLine("imap host: ")
	if err != nil {
		return err
	}

	line, err := t.ReadLine("imap port: ")
	if err != nil {
		return err
	}

	port, err := strconv.ParseInt(line, 10, 16)
	if err != nil {
		return err
	}
	p.ImapPort = uint16(port)
	p.ImapUser, err = t.ReadLine("imap Username: ")
	if err != nil {
		return err
	}

	pwd, err := t.ReadPass("imap Password: ")
	if err != nil {
		return fmt.Errorf("cannot read imap password.")
	}
	c := cred.New()
	err = c.Add(p.ImapUser, "imap", p.ImapHost, p.ImapPort, string(pwd))
	if err != nil {
		prompt := fmt.Sprintf("credentials for imap://%s@%s:%d already exist.\ndo you want to provide a new password? [y/n]: ", p.ImapUser, p.ImapHost, p.ImapPort)
		ans, err := t.ReadLine(prompt)
		if err != nil {
			return fmt.Errorf("cannot read answer.")
		}
		if len(ans) == 0 {
			ans = "n"
		}
		if ans[0] == 'y' || ans[0] == 'Y' {
			err = c.Update(p.ImapUser, "imap", p.ImapHost, p.ImapPort, string(pwd))
			if err != nil {
				return err
			}
		}
	}

	p.SmtpHost, err = t.ReadLine("smtp host: ")
	if err != nil {
		return err
	}

	line, err = t.ReadLine("smtp port: ")
	if err != nil {
		return err
	}

	port, err = strconv.ParseInt(line, 10, 16)
	if err != nil {
		return err
	}
	p.SmtpPort = uint16(port)
	p.SmtpUser, err = t.ReadLine("smtp Username: ")
	if err != nil {
		return err
	}

	pwd, err = t.ReadPass("smtp Password: ")
	if err != nil {
		return fmt.Errorf("cannot read smtp password.")
	}

	err = c.Add(p.SmtpUser, "smtp", p.SmtpHost, p.SmtpPort, string(pwd))
	if err != nil {
		prompt := fmt.Sprintf("credentials for smtp://%s@%s:%d already exist.\ndo you want to provide a new password? [y/n]: ", p.SmtpUser, p.SmtpHost, p.SmtpPort)
		ans, err := t.ReadLine(prompt)
		if err != nil {
			return fmt.Errorf("cannot read answer.")
		}
		if len(ans) == 0 {
			ans = "n"
		}
		if ans[0] == 'y' || ans[0] == 'Y' {
			err = c.Update(p.SmtpUser, "smtp", p.SmtpHost, p.SmtpPort, string(pwd))
			if err != nil {
				return err
			}
		}
	}
	if err != nil {
		return err
	}
	cfg.Profiles = append(cfg.Profiles, p)

	err = Generate(cfg)
	if err != nil {
		return err
	}

	return nil
}

func Generate(cfg *config.Config) error {
	err := generatemu4e(cfg)
	if err != nil {
		return err
	}

	return nil
}

//go:embed mu4e.tpl
var mu4e string

func generatemu4e(cfg *config.Config) error {
	tmpl, err := template.New("mu4e").Parse(mu4e)
	if err != nil {
		return err
	}
	var mu4e = &bytes.Buffer{}

	err = tmpl.Execute(mu4e, cfg.Profiles)
	if err != nil {
		return err

	}
	io.Write(path.Join(cfg.EmacsCfgDir, "mu4e.el"), mu4e.Bytes(), 0644)

	return nil
}
