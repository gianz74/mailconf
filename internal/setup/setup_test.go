package setup

import (
	"path"
	"testing"

	"github.com/gianz74/mailconf/internal/cred"
	"github.com/gianz74/mailconf/internal/cred/memcred"
	"github.com/gianz74/mailconf/internal/myterm"
	"github.com/gianz74/mailconf/internal/myterm/memterm"
	"github.com/gianz74/mailconf/internal/os"
	"github.com/gianz74/mailconf/internal/service/mockservice"
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
	mockservice.SetupMockServices()
	return nil
}

func restore() {
	mockservice.RestoreServices()
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
		cred    *creds
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
			nil,
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
			nil,
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
			&creds{
				"imap://test@gmail.com:secret@imap.gmail.com:997",
				"smtp://test@gmail.com:secret@smtp.gmail.com:456",
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
			&creds{
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
			fs = testutil.NewFs(testutil.Name(t.Name()), testutil.SubName(tc.name), testutil.System(system))
			os.Set(fs)
			err := setup()
			defer restore()
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
				if tc.cred != nil {
					got, want := testutil.CheckCreds(tc.cred.imappwd)
					if got != want {
						t.Fatalf("%s: got imap pwd: %s, want: %s\n", tc.name, got, want)
					}
					got, want = testutil.CheckCreds(tc.cred.smtppwd)
					if got != want {
						t.Fatalf("%s: got smtp pwd: %s, want: %s\n", tc.name, got, want)
					}
				}
			}
		}
	}
}
