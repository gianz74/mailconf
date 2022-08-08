package service

import (
	"io/ioutil"
	"path"
	"reflect"
	"testing"

	"github.com/gianz74/mailconf/internal/config"
	"github.com/gianz74/mailconf/internal/os"
	"github.com/spf13/afero"
)

var (
	oldFs os.FsAccess
)

func setup() {
	oldFs = os.Set(&afero.Afero{
		Fs: afero.NewMemMapFs(),
	})
}

func restore() {
	os.Set(oldFs)
}

func fixture(path string) []byte {
	b, err := ioutil.ReadFile("testdata/fixtures" + path)
	if err != nil {
		panic(err)
	}
	return b
}

func NewmockMbsync(cfg *config.Config, system string) Service {
	switch system {
	case "linux":
		return NewmockMbsyncLinux(cfg)
	case "darwin":
		return NewmockMbsyncDarwin(cfg)
	default:
		return nil
	}
}

func NewmockMbsyncLinux(cfg *config.Config) Service {
	return &mockMbsyncLinux{
		MbsyncLinux: MbsyncLinux{
			cfg: cfg,
		},
	}
}

type mockMbsyncLinux struct {
	MbsyncLinux
}

func (mockMbsyncLinux) Start() error {
	return nil
}

func (mockMbsyncLinux) Stop() error {
	return nil
}

func (mockMbsyncLinux) Enable() error {
	return nil
}

func (mockMbsyncLinux) Disable() error {
	return nil
}

func NewmockMbsyncDarwin(cfg *config.Config) Service {
	return &mockMbsyncDarwin{
		MbsyncDarwin: MbsyncDarwin{
			cfg: cfg,
		},
	}
}

type mockMbsyncDarwin struct {
	MbsyncDarwin
}

func (mockMbsyncDarwin) Start() error {
	return nil
}

func (mockMbsyncDarwin) Stop() error {
	return nil
}

func (mockMbsyncDarwin) Enable() error {
	return nil
}

func (mockMbsyncDarwin) Disable() error {
	return nil
}

func TestGenerateMbsync(t *testing.T) {
	tt := []struct {
		name    string
		systems []string
		force   bool
		config  *config.Config
		err     error
	}{
		{
			"single",
			[]string{
				"linux",
				"darwin",
			},
			false,
			&config.Config{
				EmacsCfgDir: "/home/user/.emacs.d",
				BinDir:      "/home/user/.local/bin",
				Profiles: []*config.Profile{
					{
						Name:     "Work",
						FullName: "John Doe",
						Email:    "jdoe@gmail.com",
						ImapHost: "imap.gmail.com",
						ImapPort: 993,
						ImapUser: "user@gmail.com",
						SmtpHost: "smtp.gmail.com",
						SmtpPort: 587,
						SmtpUser: "user@gmail.com",
					},
				},
			},
			nil,
		},
	}
	setup()
	defer restore()

	for _, tc := range tt {
		for _, system := range tc.systems {
			svc := NewmockMbsync(tc.config, system)
			err := svc.GenConf(tc.force)
			if err != tc.err {
				t.Fatalf("%s: got: %v, want: %v", tc.name, err, tc.err)
			}
			var got []byte
			var want []byte
			switch system {
			case "linux":
				cfgdir, err := os.UserConfigDir()
				if err != nil {
					t.Fatalf("cannot get user config dir: %v", err)
				}
				got, err = os.ReadFile(path.Join(cfgdir, "systemd/user/mbsync.service"))
				if err != nil {
					t.Fatalf("cannot get written file: %v\n", err)
				}
				want = fixture("/generatembsync/" + tc.name + "/" + system + "/mbsync.service")
				got, err = os.ReadFile(path.Join(cfgdir, "systemd/user/mbsync.timer"))
				if err != nil {
					t.Fatalf("cannot get written file: %v\n", err)
				}
				want = fixture("/generatembsync/" + tc.name + "/" + system + "/mbsync.timer")
			case "darwin":
				homedir, err := os.UserHomeDir()
				if err != nil {
					t.Fatalf("cannot get user home dir: %v", err)
				}
				got, err = os.ReadFile(path.Join(homedir, "Library/LaunchAgents/local.mbsync.plist"))
				if err != nil {
					t.Fatalf("cannot get written file: %v\n", err)
				}
				want = fixture("/generatembsync/" + tc.name + "/" + system + "/local.mbsync.plist")
			}
			if err != nil {
				t.Fatalf("file not saved: %v", err)

			}
			if !reflect.DeepEqual(want, got) {
				t.Fatalf("%s: got: %s, want: %s", tc.name, got, want)
			}

		}
	}
}
