package service

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"reflect"
	"regexp"
	"text/template"

	"github.com/gianz74/mailconf/internal/config"
	"github.com/gianz74/mailconf/internal/io"
	"github.com/gianz74/mailconf/internal/os"
)

var (
	ErrExists     = errors.New("config already exists")
	NewImapnotify = newImapnotify
	NewMbsync     = newMbsync
)

type Status int

func (s Status) String() string {
	return statuses[s]
}

const (
	Unknown Status = iota
	DisabledRunning
	DisabledStopped
	EnabledRunning
	EnabledStopped
	Notexistent
)

var (
	statuses = []string{
		"Unknown",
		"DisabledRunning",
		"DisabledStopped",
		"EnabledRunning",
		"EnabledStopped",
		"Notexistent",
	}
)

type Service interface {
	Start() error
	Stop() error
	Enable() error
	Disable() error
	Status() Status
	GenConf(bool) error
}

func newMbsync(cfg *config.Config) Service {
	switch os.System {
	case "linux":
		return NewMbsyncLinux(cfg)
	case "darwin":
		return NewMbsyncDarwin(cfg)
	default:
		return nil
	}
}

func newImapnotify(cfg *config.Config, profile *config.Profile) Service {
	switch os.System {
	case "linux":
		return NewImapnotifyLinux(cfg, profile)
	case "darwin":
		return NewImapnotifyDarwin(cfg, profile)
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
	cmd := exec.Command("systemctl", "--user", "start", "mbsync.timer")
	if err := cmd.Start(); err != nil {
		return err
	}
	err := cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (m MbsyncLinux) Stop() error {
	cmd := exec.Command("systemctl", "--user", "stop", "mbsync.timer")
	if err := cmd.Start(); err != nil {
		return err
	}
	err := cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (m MbsyncLinux) Enable() error {
	cmd := exec.Command("systemctl", "--user", "enable", "mbsync.timer")
	if err := cmd.Start(); err != nil {
		return err
	}
	err := cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (m MbsyncLinux) Disable() error {
	cmd := exec.Command("systemctl", "--user", "disable", "mbsync.timer")
	if err := cmd.Start(); err != nil {
		return err
	}
	err := cmd.Wait()
	if err != nil {
		return err
	}

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

func (m MbsyncLinux) Status() Status {
	cmd := exec.Command("systemctl", "--user", "status", "mbsync.timer")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Unknown
	}
	if err := cmd.Start(); err != nil {
		return Unknown
	}
	s := bufio.NewScanner(stdout)
	st := regexp.MustCompile(`\s*Active:\s*(?P<status>[^\s]*)\s.*`)
	en := regexp.MustCompile(`\s*Loaded:[^;]+;\s+(?P<enabled>[^;]*);.*`)
	result := make(map[string]string)
	for s.Scan() {
		line := s.Text()
		match := st.FindStringSubmatch(line)
		for i, name := range st.SubexpNames() {
			if i != 0 && name != "" {
				if i < len(match) {
					result[name] = match[i]
				}
			}
		}
		match = en.FindStringSubmatch(line)
		for i, name := range en.SubexpNames() {
			if i != 0 && name != "" {
				if i < len(match) {
					result[name] = match[i]
				}
			}
		}
	}
	enabled, ok := result["enabled"]
	if !ok {
		return Notexistent
	}
	status, ok := result["status"]
	if !ok {
		return Notexistent
	}
	switch status {
	case "inactive":
		if enabled == "enabled" {
			return EnabledStopped
		}
		return DisabledStopped
	case "active":
		if enabled == "enabled" {
			return EnabledRunning
		}
		return DisabledRunning
	default:
		fmt.Printf("status: %s\n", status)
		return Unknown
	}
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

func (m MbsyncDarwin) Status() Status {
	return Unknown
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
	cmd := exec.Command("systemctl", "--user", "start", fmt.Sprintf("imapnotify@%s.service", m.profile.Name))
	if err := cmd.Start(); err != nil {
		return err
	}
	err := cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (m ImapnotifyLinux) Stop() error {
	cmd := exec.Command("systemctl", "--user", "stop", fmt.Sprintf("imapnotify@%s.service", m.profile.Name))
	if err := cmd.Start(); err != nil {
		return err
	}
	err := cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (m ImapnotifyLinux) Enable() error {
	cmd := exec.Command("systemctl", "--user", "enable", fmt.Sprintf("imapnotify@%s.service", m.profile.Name))
	if err := cmd.Start(); err != nil {
		return err
	}
	err := cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (m ImapnotifyLinux) Disable() error {
	cmd := exec.Command("systemctl", "--user", "disable", fmt.Sprintf("imapnotify@%s.service", m.profile.Name))
	if err := cmd.Start(); err != nil {
		return err
	}
	err := cmd.Wait()
	if err != nil {
		return err
	}

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

func (m ImapnotifyLinux) Status() Status {
	cmd := exec.Command("systemctl", "--user", "status", fmt.Sprintf("imapnotify@%s.service", m.profile.Name))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Unknown
	}
	if err := cmd.Start(); err != nil {
		return Unknown
	}
	s := bufio.NewScanner(stdout)
	st := regexp.MustCompile(`\s*Active:\s*(?P<status>[^\s]*)\s.*`)
	en := regexp.MustCompile(`\s*Loaded:[^;]+;\s+(?P<enabled>[^;]*);.*`)
	result := make(map[string]string)
	for s.Scan() {
		line := s.Text()
		match := st.FindStringSubmatch(line)
		for i, name := range st.SubexpNames() {
			if i != 0 && name != "" {
				if i < len(match) {
					result[name] = match[i]
				}
			}
		}
		match = en.FindStringSubmatch(line)
		for i, name := range en.SubexpNames() {
			if i != 0 && name != "" {
				if i < len(match) {
					result[name] = match[i]
				}
			}
		}
	}
	enabled, ok := result["enabled"]
	if !ok {
		return Notexistent
	}
	status, ok := result["status"]
	if !ok {
		return Notexistent
	}
	switch status {
	case "inactive":
		if enabled == "enabled" {
			return EnabledStopped
		}
		return DisabledStopped
	case "active":
		if enabled == "enabled" {
			return EnabledRunning
		}
		return DisabledRunning
	default:
		fmt.Printf("status: %s\n", status)
		return Unknown
	}
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

func (m ImapnotifyDarwin) Status() Status {
	return Unknown
}
