package add

import (
	"testing"

	"github.com/gianz74/mailconf"
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

type creds struct {
	imappwd string
	smtppwd string
}

func setup() error {
	myterm.SetTerm(mockTerm)
	os.UserConfigDir = func() (string, error) { return "/home/user/.config", nil }
	os.UserHomeDir = func() (string, error) { return "/home/user", nil }
	cred.SetStore(mockCredStore)
	return nil
}

func TestAdd(t *testing.T) {
	tt := []struct {
		name        string
		systems     []string
		creds       []string
		chat        []string
		expectCreds *creds
		err         error
	}{
		{
			"NoConfig",
			[]string{
				"linux",
				"darwin",
			},
			[]string{},
			[]string{
				"Test",
			},
			nil,
			ErrNoConfig,
		},
		{
			"ProfileExists",
			[]string{
				"linux",
				"darwin",
			},
			[]string{},
			[]string{
				"Test",
			},
			nil,
			mailconf.ErrProfileExists,
		},
		{
			"FirstProfile",
			[]string{
				"linux",
				"darwin",
			},
			[]string{},
			[]string{
				"OldProfile",
				"John Doe the elder",
				"jdoe_old@gmail.com",
				"imap.gmail.com",
				"997",
				"jdoe_old@gmail.com",
				"oldimapsecret",
				"smtp.gmail.com",
				"456",
				"jdoe_old@gmail.com",
				"oldsmtpsecret",
			},
			&creds{
				"imap://jdoe_old@gmail.com:oldimapsecret@imap.gmail.com:997",
				"smtp://jdoe_old@gmail.com:oldsmtpsecret@smtp.gmail.com:456",
			},
			nil,
		},
		{
			"NewProfile",
			[]string{
				"linux",
				"darwin",
			},
			[]string{},
			[]string{
				"Test",
				"John Doe",
				"jdoe@gmail.com",
				"imap.gmail.com",
				"997",
				"jdoe@gmail.com",
				"newimapsecret",
				"smtp.gmail.com",
				"456",
				"jdoe@gmail.com",
				"newsmtpsecret",
			},
			&creds{
				"imap://jdoe@gmail.com:newimapsecret@imap.gmail.com:997",
				"smtp://jdoe@gmail.com:newsmtpsecret@smtp.gmail.com:456",
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

			mockTerm.SetLines(tc.chat)
			mockCredStore.AddBulk(tc.creds)
			err = runAdd(nil, nil)
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
				if tc.expectCreds != nil {
					got, want := testutil.CheckCreds(tc.expectCreds.imappwd)
					if got != want {
						t.Fatalf("%s: got imap pwd: %s, want: %s\n", tc.name, got, want)
					}
					got, want = testutil.CheckCreds(tc.expectCreds.smtppwd)
					if got != want {
						t.Fatalf("%s: got smtp pwd: %s, want: %s\n", tc.name, got, want)
					}
				}
			}

		}
	}
}
