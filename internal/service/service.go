package service

import (
	"bytes"
	_ "embed"
	"errors"
	"path"
	"reflect"
	"runtime"
	"text/template"

	"github.com/gianz74/mailconf/internal/config"
	"github.com/gianz74/mailconf/internal/io"
	"github.com/gianz74/mailconf/internal/os"
)

var (
	NewMbsync     = NewMbsyncOs(runtime.GOOS)
	NewImapnotify = NewImapnotifyOs(runtime.GOOS)
	ErrExists     = errors.New("config already exists")
)

type Service interface {
	Start() error
	Stop() error
	Enable() error
	Disable() error
	GenConf(bool) error
}

func NewMbsyncOs(OS string) func(*config.Config) Service {
	switch OS {
	case "linux":
		return NewMbsyncLinux
	case "darwin":
		return NewMbsyncDarwin
	default:
		return nil
	}
}

func NewImapnotifyOs(OS string) func(*config.Config, *config.Profile) Service {
	switch OS {
	case "linux":
		return NewImapnotifyLinux
	case "darwin":
		return NewImapnotifyDarwin
	default:
		return nil
	}
}

func NewMbsyncLinux(cfg *config.Config) Service {
	return MbsyncLinux{
		cfg: cfg,
	}
}

func NewMbsyncDarwin(cfg *config.Config) Service {
	return MbsyncDarwin{
		cfg: cfg,
	}
}

type MbsyncLinux struct {
	cfg *config.Config
}

func (m MbsyncLinux) Start() error {
	return nil
}

func (m MbsyncLinux) Stop() error {
	return nil
}

func (m MbsyncLinux) Enable() error {
	return nil
}

func (m MbsyncLinux) Disable() error {
	return nil
}

//go:embed templates/linux/mbsync.service.tmpl
var mbsyncsvclinux string

//go:embed templates/linux/mbsync.timer.tmpl
var mbsynctimerlinux []byte

func (m MbsyncLinux) GenConf(force bool) error {
	tmpl, err := template.New("mbsync.service").Parse(mbsyncsvclinux)
	if err != nil {
		return err
	}
	var mbsyncsvc = &bytes.Buffer{}

	err = tmpl.Execute(mbsyncsvc, m.cfg)
	if err != nil {
		return err

	}
	cfgdir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	tmp, err := os.ReadFile(path.Join(cfgdir, "systemd/user/mbsync.timer"))
	if err == nil && !(reflect.DeepEqual(tmp, mbsynctimerlinux) || force) {
		return ErrExists
	}
	tmp, err = os.ReadFile(path.Join(cfgdir, "systemd/user/mbsync.service"))
	if err == nil && !(reflect.DeepEqual(tmp, mbsyncsvc.Bytes()) || force) {
		return ErrExists
	}

	io.Write(path.Join(cfgdir, "systemd/user/mbsync.timer"), mbsynctimerlinux, 0644)
	io.Write(path.Join(cfgdir, "systemd/user/mbsync.service"), mbsyncsvc.Bytes(), 0644)

	return nil
}

type MbsyncDarwin struct {
	cfg *config.Config
}

func (m MbsyncDarwin) Start() error {
	return nil
}

func (m MbsyncDarwin) Stop() error {
	return nil
}

func (m MbsyncDarwin) Enable() error {
	return nil
}

func (m MbsyncDarwin) Disable() error {
	return nil
}

//go:embed templates/darwin/local.mbsync.plist.tmpl
var mbsyncsvcdarwin string

func (m MbsyncDarwin) GenConf(force bool) error {

	tmpl, err := template.New("local.mbsync.plist").Parse(mbsyncsvcdarwin)
	if err != nil {
		return err
	}
	var mbsyncsvc = &bytes.Buffer{}

	err = tmpl.Execute(mbsyncsvc, m.cfg)
	if err != nil {
		return err

	}
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	tmp, err := os.ReadFile(path.Join(homedir, "Library/LaunchAgents/local.mbsync.plist"))
	if err == nil && !(reflect.DeepEqual(tmp, mbsyncsvc.Bytes()) || force) {
		return ErrExists
	}

	err = io.Write(path.Join(homedir, "Library/LaunchAgents/local.mbsync.plist"), mbsyncsvc.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func NewImapnotifyLinux(cfg *config.Config, profile *config.Profile) Service {
	return ImapnotifyLinux{
		cfg:     cfg,
		profile: profile,
	}
}

func NewImapnotifyDarwin(cfg *config.Config, profile *config.Profile) Service {
	return ImapnotifyDarwin{
		cfg:     cfg,
		profile: profile,
	}
}

type ImapnotifyLinux struct {
	cfg     *config.Config
	profile *config.Profile
}

func (m ImapnotifyLinux) Start() error {
	return nil
}

func (m ImapnotifyLinux) Stop() error {
	return nil
}

func (m ImapnotifyLinux) Enable() error {
	return nil
}

func (m ImapnotifyLinux) Disable() error {
	return nil
}

//go:embed templates/linux/imapnotify.service.tmpl
var imapnotifysvclinux string

func (m ImapnotifyLinux) GenConf(force bool) error {
	tmpl, err := template.New("imapnotify.service").Parse(imapnotifysvclinux)
	if err != nil {
		return err
	}
	var imapnotifysvc = &bytes.Buffer{}

	err = tmpl.Execute(imapnotifysvc, m.cfg)
	if err != nil {
		return err

	}
	cfgdir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	tmp, err := os.ReadFile(path.Join(cfgdir, "systemd/user/imapnotify@.service"))
	if err == nil && !(reflect.DeepEqual(tmp, imapnotifysvc.Bytes()) || force) {
		return ErrExists
	}

	io.Write(path.Join(cfgdir, "systemd/user/imapnotify@.service"), imapnotifysvc.Bytes(), 0644)

	return nil
}

type ImapnotifyDarwin struct {
	cfg     *config.Config
	profile *config.Profile
}

func (m ImapnotifyDarwin) Start() error {
	return nil
}

func (m ImapnotifyDarwin) Stop() error {
	return nil
}

func (m ImapnotifyDarwin) Enable() error {
	return nil
}

func (m ImapnotifyDarwin) Disable() error {
	return nil
}

//go:embed templates/darwin/imapnotify.plist.tmpl
var imapnotifysvcdarwin string

func (m ImapnotifyDarwin) GenConf(force bool) error {
	tmpl, err := template.New("imapnotify.service").Parse(imapnotifysvcdarwin)
	if err != nil {
		return err
	}
	var imapnotifysvc = &bytes.Buffer{}

	cfgdir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	param := struct {
		Cfg     *config.Config
		Profile *config.Profile
		CfgDir  string
	}{
		Cfg:     m.cfg,
		Profile: m.profile,
		CfgDir:  cfgdir,
	}

	err = tmpl.Execute(imapnotifysvc, param)
	if err != nil {
		return err

	}
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	tmp, err := os.ReadFile(path.Join(homedir, "Library/LaunchAgents/local.imapnotify."+m.profile.Name+".plist"))
	if err == nil && !(reflect.DeepEqual(tmp, imapnotifysvc.Bytes()) || force) {
		return ErrExists
	}

	io.Write(path.Join(homedir, "Library/LaunchAgents/local.imapnotify."+m.profile.Name+".plist"), imapnotifysvc.Bytes(), 0644)

	return nil
}
