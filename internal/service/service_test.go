package service

import (
	"io/ioutil"
	"testing"

	"github.com/gianz74/mailconf/internal/config"
	"github.com/gianz74/mailconf/internal/options"
	"github.com/gianz74/mailconf/internal/os"
	"github.com/gianz74/mailconf/internal/testutil"
	"github.com/spf13/afero"
)

var (
	oldFs os.FsAccess
)

func setup() {
	os.UserConfigDir = func() (string, error) { return "/home/user/.config", nil }
	os.UserHomeDir = func() (string, error) { return "/home/user", nil }
	setupMockServices()
}

func restore() {
	restoreServices()
}

func fixture(path string) []byte {
	b, err := ioutil.ReadFile("testdata/fixtures" + path)
	if err != nil {
		panic(err)
	}
	return b
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

	for _, tc := range tt {
		for _, system := range tc.systems {
			setup()
			defer restore()
			var fs afero.Afero
			os.System = system
			fs = testutil.NewFs(testutil.Name(t.Name()), testutil.SubName(tc.name))
			os.Set(fs)
			svc := NewMbsync(tc.config)
			err := svc.GenConf(tc.force)
			if err != tc.err {
				t.Fatalf("%s: got: %v, want: %v", tc.name, err, tc.err)
			}
			for file := range testutil.GetFiles(t.Name(), tc.name, system) {
				got, err := testutil.Result(file)
				if err != nil {
					t.Fatalf("%s (%s): missing file %s\n", tc.name, system, file)
				}

				want := testutil.Fixture(t.Name(), tc.name, system, file)
				if got != want {
					t.Fatalf("%s: got: %s, want: %s\n", tc.name, got, want)
				}
			}
		}
	}
}

func TestRemoveMbsync(t *testing.T) {
	tt := []struct {
		name    string
		systems []string
		err     error
	}{
		{
			"single",
			[]string{
				"linux",
				"darwin",
			},
			nil,
		},
	}

	for _, tc := range tt {
		for _, system := range tc.systems {
			setup()
			defer restore()
			var fs afero.Afero
			os.System = system
			fs = testutil.NewFs(testutil.Name(t.Name()), testutil.SubName(tc.name), testutil.System(system))
			os.Set(fs)
			cfg := config.Read()
			svc := NewMbsync(cfg)
			options.Set(options.OptVerbose(true))
			err := svc.Remove()
			if err != tc.err {
				t.Fatalf("%s: got: %v, want: %v", tc.name, err, tc.err)
			}
			for file := range testutil.GetFiles(t.Name(), tc.name, system) {
				_, err = testutil.Result(file)
				if err == nil {
					t.Fatalf("%s (%s): file %s not removed\n", tc.name, system, file)
				}
			}
		}
	}
}

func TestGenerateimapnotify(t *testing.T) {
	tt := []struct {
		name    string
		systems []string
		profile *config.Profile
		err     error
	}{
		{
			"single",
			[]string{
				"linux",
				"darwin",
			},
			&config.Profile{
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
			nil,
		},
	}
	for _, tc := range tt {
		for _, system := range tc.systems {
			setup()
			defer restore()
			var fs afero.Afero
			os.System = system
			fs = testutil.NewFs(testutil.Name(t.Name()), testutil.SubName(tc.name), testutil.System(system))
			os.Set(fs)
			err := generateimapnotify(tc.profile, false)
			if err != nil {
				t.Fatalf("cannot generate file: %v", err)
			}
			for file := range testutil.GetFiles(t.Name(), tc.name, system) {
				got, err := testutil.Result(file)
				if err != nil {
					t.Fatalf("%s (%s): missing file %s\n", tc.name, system, file)
				}

				want := testutil.Fixture(t.Name(), tc.name, system, file)
				if got != want {
					t.Fatalf("%s: got: %s, want: %s\n", tc.name, got, want)
				}
			}
		}
	}
}

func TestGeneratembsyncrc(t *testing.T) {
	tt := []struct {
		name    string
		systems []string
		err     error
	}{
		{
			"single",
			[]string{
				"linux",
				"darwin",
			},
			nil,
		},
		{
			"two_profiles",
			[]string{
				"linux",
				"darwin",
			},
			nil,
		},
	}
	for _, tc := range tt {
		for _, system := range tc.systems {
			setup()
			defer restore()
			var fs afero.Afero
			os.System = system
			fs = testutil.NewFs(testutil.Name(t.Name()), testutil.SubName(tc.name), testutil.System(system))
			os.Set(fs)
			cfg := config.Read()

			err := generatembsyncrc(cfg, false)
			if err != nil {
				t.Fatalf("cannot generate file: %v", err)
			}

			for file := range testutil.GetFiles(t.Name(), tc.name, system) {
				got, err := testutil.Result(file)
				if err != nil {
					t.Fatalf("%s (%s): missing file %s\n", tc.name, system, file)
				}

				want := testutil.Fixture(t.Name(), tc.name, system, file)
				if got != want {
					t.Fatalf("%s: got: %s, want: %s\n", tc.name, got, want)
				}
			}
		}
	}
}

func TestGenerateimapfilter(t *testing.T) {
	tt := []struct {
		name    string
		systems []string
		err     error
	}{
		{
			"single",
			[]string{
				"linux",
				"darwin",
			},
			nil,
		},
		{
			"two_profiles",
			[]string{
				"linux",
				"darwin",
			},
			nil,
		},
	}
	for _, tc := range tt {
		for _, system := range tc.systems {
			setup()
			defer restore()
			var fs afero.Afero
			os.System = system
			fs = testutil.NewFs(testutil.Name(t.Name()), testutil.SubName(tc.name), testutil.System(system))
			os.Set(fs)
			cfg := config.Read()

			err := generateimapfilter(cfg, false)
			if err != nil {
				t.Fatalf("cannot generate file: %v", err)
			}

			for file := range testutil.GetFiles(t.Name(), tc.name, system) {
				got, err := testutil.Result(file)
				if err != nil {
					t.Fatalf("%s (%s): missing file %s\n", tc.name, system, file)
				}

				want := testutil.Fixture(t.Name(), tc.name, system, file)
				if got != want {
					t.Fatalf("%s: got: %s, want: %s\n", tc.name, got, want)
				}
			}
		}
	}
}

func setupMockServices() {
	SetMbsync(newMockMbsync)
	SetImapnotify(newMockImapnotify)
}

func restoreServices() {
	SetMbsync(MbsyncCtor)
	SetImapnotify(ImapnotifyCtor)
}

func newMockMbsync(cfg *config.Config) Service {
	return &MockService{
		Service: MbsyncCtor(cfg),
	}
}

func newMockImapnotify(cfg *config.Config, profile *config.Profile) Service {
	return &MockService{
		Service: ImapnotifyCtor(cfg, profile),
	}
}

type MockService struct {
	Service
}

func (MockService) Start()   {}
func (MockService) Stop()    {}
func (MockService) Enable()  {}
func (MockService) Disable() {}
