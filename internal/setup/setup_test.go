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
	"github.com/gianz74/mailconf/internal/testutil"
	"github.com/spf13/afero"
)

var (
	mockTerm      = memterm.New()
	mockCredStore = memcred.New()
)

func setup() error {
	myterm.SetTerm(mockTerm)
	checkRequirements = func(string) bool { return true }
	os.UserConfigDir = func() (string, error) { return "/home/user/.config", nil }
	os.UserHomeDir = func() (string, error) { return "/home/user", nil }
	cred.SetStore(mockCredStore)
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

func readConf() (string, error) {
	cfgdir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	cfgFile := path.Join(cfgdir, "mailconf", "data.json")
	out, err := os.ReadFile(cfgFile)
	return string(out), err
}

type creds struct {
	imappwd string
	smtppwd string
}

func TestSetup(t *testing.T) {
	tt := []struct {
		name    string
		systems []string
		creds   []string
		input   []string
		cred    creds
		err     error
	}{

		{
			"ConfExists",
			[]string{
				"linux",
				"darwin",
			},
			[]string{},
			[]string{},
			creds{
				"",
				"",
			},
			ErrExists,
		},
		{
			"ConfNotExisting",
			[]string{
				"linux",
				"darwin",
			},
			[]string{},
			[]string{"~/.emacs.d",
				"~/.local/bin",
				"n",
			},
			creds{
				"",
				"",
			},

			nil,
		},
		{
			"CreateProfile",
			[]string{
				"linux",
				"darwin",
			},
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
			creds{
				"secret",
				"secret",
			},
			nil,
		},
		{
			"CreateProfile Existing Credentials",
			[]string{
				"linux",
				"darwin",
			},
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
			creds{
				"imap://test@gmail.com:newsecret@imap.gmail.com:997",
				"smtp://test@gmail.com:secret@smtp.gmail.com:456",
			},
			nil,
		},
	}
	for _, tc := range tt {
		for _, system := range tc.systems {
			var fs afero.Afero
			os.System = system
			fs = testutil.NewFs(testutil.Name(tc.name), testutil.System(system))
			os.Set(fs)
			err := setup()
			if err != nil {
				t.Fatalf("%s: cannot prepare environment: %v\n", tc.name, err)
			}

			mockTerm.SetLines(tc.input)
			mockCredStore.AddBulk(tc.creds)
			err = runSetup(nil, []string{})
			if tc.err != err {
				t.Fatalf("%s: got error %v, want: %v\n", tc.name, err, tc.err)
			}
			if tc.err == nil {
				for file := range testutil.GetFiles(tc.name, system) {
					got := testutil.Result(file)
					want := testutil.Fixture(tc.name, system, file)
					if got != want {
						t.Fatalf("%s: got: %s, want: %s\n", tc.name, got, want)
					}
				}
				got, want := checkCreds(tc.cred.imappwd)
				if got != want {
					t.Fatalf("%s: got imap pwd: %s, want: %s\n", tc.name, got, want)
				}
				got, want = checkCreds(tc.cred.smtppwd)
				if got != want {
					t.Fatalf("%s: got smtp pwd: %s, want: %s\n", tc.name, got, want)
				}
			}
		}
	}
}
