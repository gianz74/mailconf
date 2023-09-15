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
	"strings"
	"text/template"

	"github.com/gianz74/mailconf/internal/config"
	"github.com/gianz74/mailconf/internal/io"
	"github.com/gianz74/mailconf/internal/options"
	"github.com/gianz74/mailconf/internal/os"
)

type (
	CtorMbsync     func(*config.Config) Service
	CtorImapnotify func(*config.Config, *config.Profile) Service
)

var (
	ErrExists                     = errors.New("config already exists")
	_newImapnotify CtorImapnotify = ImapnotifyCtor
	_newMbsync     CtorMbsync     = MbsyncCtor
)

type Status int

const (
	Unknown Status = iota
	DisabledRunning
	DisabledStopped
	EnabledRunning
	EnabledStopped
	NotFound
)

func (s Status) String() string {
	statuses := []string{
		"Unknown",
		"DisabledRunning",
		"DisabledStopped",
		"EnabledRunning",
		"EnabledStopped",
		"NotFound",
	}
	return statuses[s]
}

func SetMbsync(f CtorMbsync) {
	_newMbsync = f
}

func SetImapnotify(f CtorImapnotify) {
	_newImapnotify = f
}

func NewMbsync(cfg *config.Config) Service {
	return _newMbsync(cfg)
}

func NewImapnotify(cfg *config.Config, profile *config.Profile) Service {
	return _newImapnotify(cfg, profile)
}

type Service interface {
	Start()
	Stop()
	Enable()
	Disable()
	Remove() error
	Status() Status
	GenConf(bool) error
}

func MbsyncCtor(cfg *config.Config) Service {
	switch os.System {
	case "linux":
		return newMbsyncLinux(cfg)
	case "darwin":
		return newMbsyncDarwin(cfg)
	default:
		return nil
	}
}

func ImapnotifyCtor(cfg *config.Config, profile *config.Profile) Service {
	switch os.System {
	case "linux":
		return newImapnotifyLinux(cfg, profile)
	case "darwin":
		return newImapnotifyDarwin(cfg, profile)
	default:
		return nil
	}
}

func newMbsyncLinux(cfg *config.Config) Service {
	return mbsyncLinux{
		cfg: cfg,
	}
}

func newMbsyncDarwin(cfg *config.Config) Service {
	return mbsyncDarwin{
		cfg: cfg,
	}
}

type mbsyncLinux struct {
	cfg *config.Config
}

func (m mbsyncLinux) Start() {
	if options.Dryrun() {
		fmt.Fprintf(os.Stdout, "starting mbsync service\n")
		return
	}
	if options.Verbose() {
		fmt.Fprintf(os.Stdout, "starting mbsync service\n")
	}
	cmd := exec.Command("systemctl", "--user", "start", "mbsync.timer")
	cmd.Start()
}

func (m mbsyncLinux) Stop() {
	if options.Dryrun() {
		fmt.Fprintf(os.Stdout, "stopping mbsync service\n")
		return
	}
	if options.Verbose() {
		fmt.Fprintf(os.Stdout, "stopping mbsync service\n")
	}
	cmd := exec.Command("systemctl", "--user", "stop", "mbsync.timer")
	cmd.Start()
}

func (m mbsyncLinux) Enable() {
	if options.Dryrun() {
		fmt.Fprintf(os.Stdout, "enabling mbsync service\n")
		return
	}
	if options.Verbose() {
		fmt.Fprintf(os.Stdout, "enabling mbsync service\n")
	}
	cmd := exec.Command("systemctl", "--user", "enable", "mbsync.timer")
	cmd.Start()
}

func (m mbsyncLinux) Disable() {
	if options.Dryrun() {
		fmt.Fprintf(os.Stdout, "disabling mbsync service\n")
		return
	}
	if options.Verbose() {
		fmt.Fprintf(os.Stdout, "disabling mbsync service\n")
	}
	cmd := exec.Command("systemctl", "--user", "disable", "mbsync.timer")
	cmd.Start()
}

func (m mbsyncLinux) Remove() error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	remove(path.Join(homedir, ".mbsyncrc"))
	remove(path.Join(homedir, ".imapfilter"))

	cfgdir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	remove(path.Join(cfgdir, "systemd/user/mbsync.service"))
	remove(path.Join(cfgdir, "systemd/user/mbsync.timer"))
	return nil
}

func remove(file string) {
	if options.Dryrun() {
		fmt.Fprintf(os.Stdout, "removing %s\n", file)
	} else {
		if options.Verbose() {
			fmt.Fprintf(os.Stdout, "removing %s\n", file)
		}
		os.RemoveAll(file)
	}
}

//go:embed templates/linux/mbsync.service.tmpl
var mbsyncsvclinux string

//go:embed templates/linux/mbsync.timer.tmpl
var mbsynctimerlinux []byte

func (m mbsyncLinux) GenConf(force bool) error {
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

	if options.Verbose() {
		fmt.Fprintf(os.Stdout, "generating .mbsyncrc")
	}

	err = generatembsyncrc(m.cfg, force)
	if err != nil {
		return err
	}

	if options.Verbose() {
		fmt.Fprintf(os.Stdout, "generating imapfilter config")
	}

	err = generateimapfilter(m.cfg, force)
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

func (m mbsyncLinux) Status() Status {
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
		return NotFound
	}
	status, ok := result["status"]
	if !ok {
		return NotFound
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

type mbsyncDarwin struct {
	cfg *config.Config
}

func (m mbsyncDarwin) Start() {
}

func (m mbsyncDarwin) Stop() {
}

func (m mbsyncDarwin) Enable() {
}

func (m mbsyncDarwin) Disable() {
}

func (m mbsyncDarwin) Remove() error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	remove(path.Join(homedir, ".mbsyncrc"))
	remove(path.Join(homedir, ".imapfilter"))

	remove(path.Join(homedir, "Library/LaunchAgents/local.mbsync.plist"))
	return nil
}

//go:embed templates/darwin/local.mbsync.plist.tmpl
var mbsyncsvcdarwin string

func (m mbsyncDarwin) GenConf(force bool) error {

	err := generatembsyncrc(m.cfg, force)
	if err != nil {
		return err
	}

	err = generateimapfilter(m.cfg, force)
	if err != nil {
		return err
	}

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

func (m mbsyncDarwin) Status() Status {
	return EnabledRunning
}

func newImapnotifyLinux(cfg *config.Config, profile *config.Profile) Service {
	return imapnotifyLinux{
		cfg:     cfg,
		profile: profile,
	}
}

func newImapnotifyDarwin(cfg *config.Config, profile *config.Profile) Service {
	return imapnotifyDarwin{
		cfg:     cfg,
		profile: profile,
	}
}

type imapnotifyLinux struct {
	cfg     *config.Config
	profile *config.Profile
}

func (m imapnotifyLinux) Start() {
	cmd := exec.Command("systemctl", "--user", "start", fmt.Sprintf("imapnotify@%s.service", m.profile.Name))
	cmd.Start()
}

func (m imapnotifyLinux) Stop() {
	cmd := exec.Command("systemctl", "--user", "stop", fmt.Sprintf("imapnotify@%s.service", m.profile.Name))
	cmd.Start()
}

func (m imapnotifyLinux) Enable() {
	cmd := exec.Command("systemctl", "--user", "enable", fmt.Sprintf("imapnotify@%s.service", m.profile.Name))
	cmd.Start()
}

func (m imapnotifyLinux) Disable() {
	cmd := exec.Command("systemctl", "--user", "disable", fmt.Sprintf("imapnotify@%s.service", m.profile.Name))
	cmd.Start()
}

func (m imapnotifyLinux) Remove() error {
	cfgdir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	remove(path.Join(cfgdir, "imapnotify/"+m.profile.Name))

	if len(m.cfg.Profiles) == 0 {
		remove(path.Join(cfgdir, "systemd/user/imapnotify@.service"))
	}
	return nil
}

//go:embed templates/imapnotify/notify.conf.tmpl
var imapnotify string

func generateimapnotify(profile *config.Profile, force bool) error {
	tmpl, err := template.New("imapnotify").Parse(imapnotify)
	if err != nil {
		return err
	}

	param := struct {
		OS      string
		Profile *config.Profile
	}{
		OS:      os.System,
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

	tmp, err := os.ReadFile(path.Join(cfgdir, "imapnotify/"+profile.Name+"/notify.conf"))
	same := reflect.DeepEqual(tmp, imapnotify.Bytes())

	if err == nil && !(same || force) {
		return ErrExists
	}
	if !same || force {
		io.Write(path.Join(cfgdir, "imapnotify/"+profile.Name+"/notify.conf"), imapnotify.Bytes(), 0644)
	}

	return nil
}

//go:embed templates/linux/imapnotify.service.tmpl
var imapnotifysvclinux string

func genimapnotifysvclinux(cfg *config.Config, profile *config.Profile, force bool) error {
	tmpl, err := template.New("imapnotify.service").Parse(imapnotifysvclinux)
	if err != nil {
		return err
	}
	var imapnotifysvc = &bytes.Buffer{}

	err = tmpl.Execute(imapnotifysvc, cfg)
	if err != nil {
		return err

	}
	cfgdir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	tmp, err := os.ReadFile(path.Join(cfgdir, "systemd/user/imapnotify@.service"))
	same := reflect.DeepEqual(tmp, imapnotifysvc.Bytes())
	if err == nil && !(same || force) {
		return ErrExists
	}
	if !same || force {
		io.Write(path.Join(cfgdir, "systemd/user/imapnotify@.service"), imapnotifysvc.Bytes(), 0644)
	}
	return nil
}

func (m imapnotifyLinux) GenConf(force bool) error {

	err := genimapnotifysvclinux(m.cfg, m.profile, force)
	if err != nil {
		return err
	}

	err = generateimapnotify(m.profile, force)
	if err != nil {
		return err
	}

	return nil
}

func (m imapnotifyLinux) Status() Status {
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
		return NotFound
	}
	status, ok := result["status"]
	if !ok {
		return NotFound
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

type imapnotifyDarwin struct {
	cfg     *config.Config
	profile *config.Profile
}

func (m imapnotifyDarwin) Start() {
}

func (m imapnotifyDarwin) Stop() {
}

func (m imapnotifyDarwin) Enable() {
}

func (m imapnotifyDarwin) Disable() {
}

func (m imapnotifyDarwin) Remove() error {
	cfgdir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	remove(path.Join(cfgdir, "imapnotify/"+m.profile.Name))

	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	remove(path.Join(homedir, "Library/LaunchAgents/local.imapnotify."+m.profile.Name+".plist"))
	return nil
}

func (m imapnotifyDarwin) GenConf(force bool) error {

	err := genimapnotifysvcdarwin(m.cfg, m.profile, force)
	if err != nil {
		return err
	}

	err = generateimapnotify(m.profile, force)
	if err != nil {
		return err
	}

	return nil
}

func (m imapnotifyDarwin) Status() Status {
	return EnabledRunning
}

//go:embed templates/darwin/imapnotify.plist.tmpl
var imapnotifysvcdarwin string

func genimapnotifysvcdarwin(cfg *config.Config, profile *config.Profile, force bool) error {
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
		Cfg:     cfg,
		Profile: profile,
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

	tmp, err := os.ReadFile(path.Join(homedir, "Library/LaunchAgents/local.imapnotify."+profile.Name+".plist"))
	if err == nil && !(reflect.DeepEqual(tmp, imapnotifysvc.Bytes()) || force) {
		return ErrExists
	}

	io.Write(path.Join(homedir, "Library/LaunchAgents/local.imapnotify."+profile.Name+".plist"), imapnotifysvc.Bytes(), 0644)

	return nil
}

//go:embed templates/mbsyncrc.tpl
var mbsyncrc string

func generatembsyncrc(cfg *config.Config, force bool) error {
	tmpl, err := template.New("mbsyncrc").Parse(mbsyncrc)
	if err != nil {
		return err
	}
	var mbsyncrc = &bytes.Buffer{}
	param := struct {
		OS       string
		Profiles []*config.Profile
	}{
		OS:       os.System,
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

	tmp, err := os.ReadFile(path.Join(home, ".mbsyncrc"))
	if err == nil && !(reflect.DeepEqual(tmp, mbsyncrc.Bytes()) || force) {
		return ErrExists
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

func generateimapfilter(cfg *config.Config, force bool) error {
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
		OS:       os.System,
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

	tmp, err := os.ReadFile(path.Join(home, ".imapfilter/certificates"))
	if err == nil && !(reflect.DeepEqual(tmp, certificates) || force) {
		return ErrExists
	}

	io.Write(path.Join(home, ".imapfilter/certificates"), certificates, 0644)

	tmp, err = os.ReadFile(path.Join(home, ".imapfilter/config.lua"))
	if err == nil && !(reflect.DeepEqual(tmp, configLua.Bytes()) || force) {
		return ErrExists
	}

	io.Write(path.Join(home, ".imapfilter/config.lua"), configLua.Bytes(), 0644)

	return nil
}
