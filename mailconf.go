package mailconf

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"path"
	"reflect"
	"strconv"

	"github.com/gianz74/mailconf/internal/config"
	"github.com/gianz74/mailconf/internal/cred"
	"github.com/gianz74/mailconf/internal/io"
	"github.com/gianz74/mailconf/internal/myterm"
	"github.com/gianz74/mailconf/internal/os"
	"github.com/gianz74/mailconf/internal/service"
)

var (
	ErrProfileExists           = errors.New("Profile exists")
	ErrModified                = errors.New("Config modified externally")
	ErrProfileNotFound         = errors.New("Profile not found")
	ErrOsNotSupported          = errors.New("OS not supported")
	ErrMbsyncStatusUnknown     = errors.New("Mbsync: unknown status")
	ErrMbsyncNotFound          = errors.New("Mbsync: Service not found")
	ErrImapnotifyStatusUnknown = errors.New("Imapnotify: unknown status")
	ErrImapnotifyNotFound      = errors.New("Imapnotify: Service not found")
)

func AddProfile(profile string, cfg *config.Config) error {

	if isConfModified(cfg) {
		t := myterm.New()
		yes := t.YesNo("Configuration modified by an external program. Overwrite? [y/n]: ")
		if !yes {
			return ErrModified
		}
	}

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
		fmt.Printf("%s\n", prompt)
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

	err = Generate(cfg, p)
	if err != nil {
		return err
	}

	return nil
}

func Generate(cfg *config.Config, profile *config.Profile) error {
	err := generatemu4e(cfg, true)
	if err != nil {
		return err
	}

	mbsync := service.NewMbsync(cfg)
	err = mbsync.GenConf(true)
	t := myterm.New()
	if err != nil {
		yes := t.YesNo("service file for mbsync already exists. Overwrite? [y/n]: ")
		if yes {
			mbsync.Stop()
			mbsync.Disable()
			err = mbsync.GenConf(true)
			if err != nil {
				return err
			}
		}
	}

	status := mbsync.Status()

	switch status {
	case service.DisabledStopped:
		mbsync.Enable()
		mbsync.Start()
	case service.DisabledRunning:
		mbsync.Enable()
	case service.EnabledStopped:
		mbsync.Start()
	case service.EnabledRunning:
		break
	case service.NotFound:
		return ErrMbsyncNotFound
	case service.Unknown:
		return ErrMbsyncStatusUnknown
	}

	imapnotify := service.NewImapnotify(cfg, profile)
	err = imapnotify.GenConf(true)
	if err != nil {
		yes := t.YesNo("imapnotify service file for " + profile.Name + " already exists. Overwrite? [y/n]: ")
		if yes {
			imapnotify.Stop()
			imapnotify.Disable()
			err = imapnotify.GenConf(true)
			if err != nil {
				return err
			}
		}
	}
	status = imapnotify.Status()

	switch status {
	case service.DisabledStopped:
		imapnotify.Enable()
		imapnotify.Start()
	case service.DisabledRunning:
		imapnotify.Enable()
	case service.EnabledStopped:
		imapnotify.Start()
	case service.EnabledRunning:
		break
	case service.NotFound:
		return ErrImapnotifyNotFound
	case service.Unknown:
		return ErrImapnotifyStatusUnknown
	}

	return nil
}

//go:embed templates/mu4e.tpl
var mu4e string

func generatemu4e(cfg *config.Config, force bool) error {
	tmpl, err := template.New("mu4e").Parse(mu4e)
	if err != nil {
		return err
	}
	var mu4e = &bytes.Buffer{}

	err = tmpl.Execute(mu4e, cfg.Profiles)
	if err != nil {
		return err

	}
	tmp, err := os.ReadFile(path.Join(cfg.EmacsCfgDir, "mu4e.el"))
	if err == nil && !(reflect.DeepEqual(tmp, mu4e.Bytes()) || force) {
		return ErrModified
	}
	io.Write(path.Join(cfg.EmacsCfgDir, "mu4e.el"), mu4e.Bytes(), 0644)

	return nil
}

func RmProfile(profile string, cfg *config.Config) error {
	var p *config.Profile
	modified := isConfModified(cfg)
	if modified {
		t := myterm.New()
		if !t.YesNo("Configuration modified by an external program. Overwrite? [y/n]: ") {
			return ErrModified
		}
	}
	mbsync := service.NewMbsync(cfg)
	err := mbsync.GenConf(true)
	if err != nil {
		return err
	}

	for idx, tmp := range cfg.Profiles {
		if profile == tmp.Name {
			p = tmp
			cfg.Profiles = append(cfg.Profiles[:idx], cfg.Profiles[idx+1:]...)
		}
	}
	if p == nil {
		return ErrProfileNotFound
	}
	imapnotifysvc := service.NewImapnotify(cfg, p)
	imapnotifysvc.Stop()
	imapnotifysvc.Disable()
	err = imapnotifysvc.Remove()
	if err != nil {
		return err
	}

	generatemu4e(cfg, true)

	if len(cfg.Profiles) == 0 {
		mbsync.Stop()
		mbsync.Disable()
		err := mbsync.Remove()
		if err != nil {
			return err
		}
		return nil
	}
	err = mbsync.GenConf(true)
	if err != nil {
		return err
	}

	return nil
}

func isConfModified(cfg *config.Config) bool {
	err := generatemu4e(cfg, false)
	if err != nil {
		return true
	}
	mbsync := service.NewMbsync(cfg)
	err = mbsync.GenConf(false)
	if err != nil {
		return true
	}

	for _, p := range cfg.Profiles {
		imapnotify := service.NewImapnotify(cfg, p)
		err := imapnotify.GenConf(false)
		if err != nil {
			return true
		}
	}
	return false
}
