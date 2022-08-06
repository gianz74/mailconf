package config

import (
	"encoding/json"
	"path"

	"github.com/gianz74/mailconf/internal/os"
)

var (
	cfile string
)

type Profile struct {
	Name     string `json:"profile_name"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	ImapHost string `json:"imaphost"`
	ImapPort uint16 `json:"imapport"`
	ImapUser string `json:"imapuser"`
	SmtpHost string `json:"smtphost"`
	SmtpPort uint16 `json:"smtpport"`
	SmtpUser string `json:"smtpuser"`
}

type Config struct {
	EmacsCfgDir string     `json:"emacs_cfg_dir"`
	BinDir      string     `json:"bindir"`
	Profiles    []*Profile `json:"profiles"`
}

func Read() *Config {
	cfgdir, err := configdir()
	if err != nil {
		return nil
	}
	di, err := os.ReadFile(path.Join(cfgdir, "data.json"))
	if err != nil {
		return nil
	}

	cfg := &Config{}

	err = json.Unmarshal(di, cfg)
	if err != nil {
		return nil
	}

	return cfg
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Save() error {
	cfgdir, err := configdir()
	if err != nil {
		return err
	}
	err = os.MkdirAll(cfgdir, os.ModePerm)
	if err != nil {
		return err
	}
	cfile := path.Join(cfgdir, "data.json")

	conf, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile(cfile, conf, 0640)
	if err != nil {
		return err
	}
	return nil
}

func configdir() (string, error) {
	cfgdir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	cfgdir = path.Join(cfgdir, "mailconf")
	return cfgdir, nil
}
