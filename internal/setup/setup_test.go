package setup

import (
	"path"
	"regexp"
	"strconv"
	"testing"

	"github.com/gianz74/mailconf/internal/cred"
	"github.com/gianz74/mailconf/internal/cred/memcred"
	"github.com/gianz74/mailconf/internal/myterm"
	"github.com/gianz74/mailconf/internal/myterm/memterm"
	"github.com/gianz74/mailconf/internal/os"
	"github.com/spf13/afero"
)

var checkRequirementsOrig func(string) bool
var (
	oldFs         os.FsAccess
	oldTerm       myterm.Terminal
	oldCredStore  cred.CredentialsStore
	mockTerm      = memterm.New()
	mockCredStore = memcred.New()
)

func setup() {
	dryrun = true
	verbose = false
	oldTerm = myterm.SetTerm(mockTerm)
	checkRequirementsOrig = checkRequirements
	checkRequirements = func(string) bool { return true }
	oldFs = os.Set(&afero.Afero{
		Fs: afero.NewMemMapFs(),
	})
	oldCredStore = cred.SetStore(mockCredStore)
}

func cleanup() {
	checkRequirements = checkRequirementsOrig
	myterm.SetTerm(oldTerm)
	cred.SetStore(oldCredStore)
	os.Set(oldFs)
}

func setupConfExists() error {
	setup()
	cfgdir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	cfgFile := path.Join(cfgdir, "mailconf", "data.json")

	os.WriteFile(cfgFile, []byte(`{
        "emacs_cfg_dir": "/home/gianz/.emacs.d",
        "bindir": "/home/gianz/.local/bin",
        "profiles": null
}`), 0644)
	return nil
}

func checkCreds(creds string) (string, string) {
	exp := regexp.MustCompile(`^(?P<service>[^:]*)://(?P<user>[^:]*):(?P<pwd>[^@]*)@(?P<host>[^:]*):(?P<port>\d+)$`)
	match := exp.FindStringSubmatch(creds)
	result := make(map[string]string)
	for i, name := range exp.SubexpNames() {
		if i != 0 && name != "" {
			if i < len(match) {
				result[name] = match[i]
			}
		}
	}
	service := result["service"]
	user := result["user"]
	pwd := result["pwd"]
	host := result["host"]
	tmp, _ := strconv.ParseUint(result["port"], 10, 16)
	port := uint16(tmp)
	got, _ := mockCredStore.Get(user, service, host, port)
	return pwd, got
}

func setupConfNotExists() error {
	setup()
	return nil
}

func readConf() (string, error) {
	cfgdir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	cfgFile := path.Join(cfgdir, "mailconf", "data.json")
	out, err := os.ReadFile(cfgFile)
	return string(out), err
}

type answer struct {
	profile string
	imappwd string
	smtppwd string
}

func TestSetup(t *testing.T) {
	tt := []struct {
		name  string
		setup func() error
		creds []string
		input []string
		want  answer
		err   error
	}{

		{
			"ConfExists",
			setupConfExists,
			[]string{},
			[]string{},
			answer{
				"",
				"",
				"",
			},
			ErrExists,
		},
		{
			"ConfNotExisting",
			setupConfNotExists,
			[]string{},
			[]string{"~/.emacs.d",
				"~/.local/bin",
				"n",
			},
			answer{
				`{
	"emacs_cfg_dir": "/home/gianz/.emacs.d",
	"bindir": "/home/gianz/.local/bin",
	"profiles": null
}`,
				"",
				"",
			},

			nil,
		},
		{
			"CreateProfile",
			setupConfNotExists,
			[]string{},
			[]string{
				"~/.emacs.d",
				"~/.local/bin",
				"y",
				"Test",
				"John Doe",
				"jdoe@gmail.com",
				"imap.gmail.com",
				"997",
				"test@gmail.com",
				"secret",
				"smtp.gmail.com",
				"456",
				"test@gmail.com",
				"secret",
				"n",
			},
			answer{`{
	"emacs_cfg_dir": "/home/gianz/.emacs.d",
	"bindir": "/home/gianz/.local/bin",
	"profiles": [
		{
			"profile_name": "Test",
			"email": "jdoe@gmail.com",
			"full_name": "John Doe",
			"imaphost": "imap.gmail.com",
			"imapport": 997,
			"imapuser": "test@gmail.com",
			"smtphost": "smtp.gmail.com",
			"smtpport": 456,
			"smtpuser": "test@gmail.com"
		}
	]
}`,
				"secret",
				"secret",
			},
			nil,
		},
		{
			"CreateProfile Existing Credentials",
			setupConfNotExists,
			[]string{
				"imap://test@gmail.com:secret@imap.gmail.com:997",
			},
			[]string{
				"~/.emacs.d",
				"~/.local/bin",
				"y",
				"Test",
				"John Doe",
				"jdoe@gmail.com",
				"imap.gmail.com",
				"997",
				"test@gmail.com",
				"newsecret",
				"y",
				"smtp.gmail.com",
				"456",
				"test@gmail.com",
				"secret",
				"n",
			},
			answer{
				`{
	"emacs_cfg_dir": "/home/gianz/.emacs.d",
	"bindir": "/home/gianz/.local/bin",
	"profiles": [
		{
			"profile_name": "Test",
			"email": "jdoe@gmail.com",
			"full_name": "John Doe",
			"imaphost": "imap.gmail.com",
			"imapport": 997,
			"imapuser": "test@gmail.com",
			"smtphost": "smtp.gmail.com",
			"smtpport": 456,
			"smtpuser": "test@gmail.com"
		}
	]
}`,
				"imap://test@gmail.com:newsecret@imap.gmail.com:997",
				"smtp://test@gmail.com:secret@smtp.gmail.com:456",
			},
			nil,
		},
	}
	for _, tc := range tt {
		err := tc.setup()
		if err != nil {
			t.Fatalf("%s: cannot prepare environment: %v\n", tc.name, err)
		}

		defer cleanup()
		mockTerm.SetLines(tc.input)
		mockCredStore.AddBulk(tc.creds)
		err = runSetup(nil, []string{})
		if tc.err != err {
			t.Fatalf("%s: got error %v, want: %v\n", tc.name, err, tc.err)
		}
		if tc.err == nil {
			got, err := readConf()
			if err != nil {
				t.Fatalf("%s: cannot read config file: %v\n", tc.name, err)
			}

			if got != tc.want.profile {
				t.Fatalf("%s: got: %s, want: %s\n", tc.name, got, tc.want)
			}
			got, want := checkCreds(tc.want.imappwd)
			if got != want {
				t.Fatalf("%s: got imap pwd: %s, want: %s\n", tc.name, got, want)
			}
			got, want = checkCreds(tc.want.smtppwd)
			if got != want {
				t.Fatalf("%s: got smtp pwd: %s, want: %s\n", tc.name, got, want)
			}
		}
	}
}
