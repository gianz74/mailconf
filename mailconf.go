package mailconf

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/gianz74/mailconf/internal/config"
	"github.com/gianz74/mailconf/internal/cred"
	"github.com/gianz74/mailconf/internal/io"
	"github.com/gianz74/mailconf/internal/myterm"
	"github.com/gianz74/mailconf/internal/os"
	"github.com/gianz74/mailconf/internal/service"
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

	err = Generate(cfg, p)
	if err != nil {
		return err
	}

	return nil
}

func Generate(cfg *config.Config, profile *config.Profile) error {
	err := generatemu4e(cfg)
	if err != nil {
		return err
	}

	err = generatembsyncrc(runtime.GOOS, cfg)
	if err != nil {
		return err
	}

	err = generateimapfilter(runtime.GOOS, cfg)
	if err != nil {
		return err
	}

	err = generateimapnotify(runtime.GOOS, profile)
	if err != nil {
		return err
	}

	mbsync := service.NewMbsync(cfg)
	err = mbsync.GenConf(false)
	t := myterm.New()
	mbsyncNew := true
	if err != nil {
		ans, err := t.ReadLine("service file for mbsync already exists. Overwrite? [y/n]: ")
		if err != nil {
			return err
		}

		if len(ans) == 0 {
			ans = "n"
		}
		if ans[0] == 'y' || ans[0] == 'Y' {
			err := mbsync.Stop()
			if err != nil {
				return err
			}

			err = mbsync.Enable()
			if err != nil {
				return err
			}

			err = mbsync.GenConf(true)
			if err != nil {
				return err
			}
		} else {
			mbsyncNew = false
		}
	}
	if mbsyncNew {
		err = mbsync.Enable()
		if err != nil {
			return err
		}

		err = mbsync.Start()
		if err != nil {
			return err
		}
	}

	imapnotify := service.NewImapnotify(cfg, profile)
	err = imapnotify.GenConf(false)
	imapnotifynew := true
	if err != nil {
		ans, err := t.ReadLine("imapnotify service file for " + profile.Name + " already exists. Overwrite? [y/n]: ")
		if err != nil {
			return err
		}

		if len(ans) == 0 {
			ans = "n"
		}
		if ans[0] == 'y' || ans[0] == 'Y' {
			err := imapnotify.Stop()
			if err != nil {
				return err
			}

			err = imapnotify.Disable()
			if err != nil {
				return err
			}

			err = imapnotify.GenConf(true)
			if err != nil {
				return err
			}
		} else {
			imapnotifynew = false
		}
	}
	if imapnotifynew {
		err = imapnotify.Enable()
		if err != nil {
			return err
		}

		err = imapnotify.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

//go:embed templates/mu4e.tpl
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

//go:embed templates/mbsyncrc.tpl
var mbsyncrc string

func generatembsyncrc(OS string, cfg *config.Config) error {
	tmpl, err := template.New("mbsyncrc").Parse(mbsyncrc)
	if err != nil {
		return err
	}
	var mbsyncrc = &bytes.Buffer{}
	param := struct {
		OS       string
		Profiles []*config.Profile
	}{
		OS:       OS,
		Profiles: cfg.Profiles,
	}

	err = tmpl.Execute(mbsyncrc, param)
	if err != nil {
		return err

	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	io.Write(path.Join(home, ".mbsyncrc"), mbsyncrc.Bytes(), 0644)

	return nil
}

//go:embed templates/imapfilter/config.lua.tmpl
var configLua string

//go:embed templates/imapfilter/certificates
var certificates []byte

func normalize(input string) string {
	fields := strings.FieldsFunc(input, func(r rune) bool {
		return r == '.' || r == '@'
	})
	return strings.Join(fields, "_")
}

func generateimapfilter(OS string, cfg *config.Config) error {
	funcMap := template.FuncMap{
		"normalize": normalize,
	}

	tmpl, err := template.New("configlua").Funcs(funcMap).Parse(configLua)
	if err != nil {
		return err
	}

	var configLua = &bytes.Buffer{}
	param := struct {
		OS       string
		Profiles []*config.Profile
	}{
		OS:       OS,
		Profiles: cfg.Profiles,
	}

	err = tmpl.Execute(configLua, param)
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	io.Write(path.Join(home, ".imapfilter/certificates"), certificates, 0644)
	io.Write(path.Join(home, ".imapfilter/config.lua"), configLua.Bytes(), 0644)

	return nil
}

//go:embed templates/imapnotify/notify.conf.tmpl
var imapnotify string

func generateimapnotify(OS string, profile *config.Profile) error {
	tmpl, err := template.New("imapnotify").Parse(imapnotify)
	if err != nil {
		return err
	}

	param := struct {
		OS      string
		Profile *config.Profile
	}{
		OS:      OS,
		Profile: profile,
	}
	imapnotify := &bytes.Buffer{}
	err = tmpl.Execute(imapnotify, param)
	if err != nil {
		return err
	}
	cfgdir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	io.Write(path.Join(cfgdir, "imapnotify/"+profile.Name+"/notify.conf"), imapnotify.Bytes(), 0644)

	return nil
}
