package mailconf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"reflect"
	"testing"

	"github.com/gianz74/mailconf/internal/config"
	"github.com/gianz74/mailconf/internal/cred"
	"github.com/gianz74/mailconf/internal/cred/memcred"
	"github.com/gianz74/mailconf/internal/myterm"
	"github.com/gianz74/mailconf/internal/myterm/memterm"
	"github.com/gianz74/mailconf/internal/os"
)

var (
	oldCredStore  cred.CredentialsStore
	oldTerm       myterm.Terminal
	mockTerm      = memterm.New()
	mockCredStore = memcred.New()
)

func setup() {
	oldTerm = myterm.SetTerm(mockTerm)
	oldCredStore = cred.SetStore(memcred.New())
	os.Set(os.MemFs)
}

func restore() {
	cred.SetStore(oldCredStore)
	myterm.SetTerm(oldTerm)
}

func fixture(path string) []byte {
	b, err := ioutil.ReadFile("testdata/fixtures" + path)
	if err != nil {
		panic(err)
	}
	return b
}

func TestAddProfile(t *testing.T) {
	tt := []struct {
		name     string
		fullname string
		email    string
		imaphost string
		imapuser string
		imapport uint16
		imappwd  string
		smtphost string
		smtpuser string
		smtpport uint16
		smtppwd  string
		conf     *config.Config
		want     *config.Config
		err      error
	}{
		{
			"Work",
			"John Doe",
			"jdoe@gmail.com",
			"imap.gmail.com",
			"user@gmail.com",
			993,
			"secret_for_imap",
			"smtp.gmail.com",
			"user@gmail.com",
			587,
			"secret_for_smtp",
			&config.Config{},
			&config.Config{
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
		{
			"Work",
			"John Doe",
			"jdoe@gmail.com",
			"asd.gmail.com",
			"asd@gmail.com",
			993,
			"secret_for_imap",
			"sdf.gmail.com",
			"sdf@gmail.com",
			587,
			"secret_for_smtp",
			&config.Config{
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
			&config.Config{
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
			ErrProfileExists,
		},
	}
	cfg := config.NewConfig()
	setup()
	defer restore()
	for _, tc := range tt {
		mockTerm.AddLine(fmt.Sprintf("%s", tc.fullname))
		mockTerm.AddLine(fmt.Sprintf("%s", tc.email))
		mockTerm.AddLine(fmt.Sprintf("%s", tc.imaphost))
		mockTerm.AddLine(fmt.Sprintf("%d", tc.imapport))
		mockTerm.AddLine(fmt.Sprintf("%s", tc.imapuser))
		mockTerm.AddLine(fmt.Sprintf("%s", tc.imappwd))
		mockTerm.AddLine(fmt.Sprintf("%s", tc.smtphost))
		mockTerm.AddLine(fmt.Sprintf("%d", tc.smtpport))
		mockTerm.AddLine(fmt.Sprintf("%s", tc.smtpuser))
		mockTerm.AddLine(fmt.Sprintf("%s", tc.smtppwd))
		err := AddProfile(tc.name, cfg)
		if err != tc.err {
			t.Fatalf("%s: got error %v, want: %v", tc.name, err, tc.err)
		}
		if !reflect.DeepEqual(cfg, tc.want) {
			w, _ := json.MarshalIndent(tc.want, "\t", "\t")
			g, _ := json.MarshalIndent(cfg, "\t", "\t")
			fmt.Printf("Want:\n%s\n", w)
			fmt.Printf("Got:\n%s\n", g)
			t.Fatalf("%s: want: %+v, got: %+v", tc.name, tc.want, cfg)
		}
	}
}

func TestGeneratemu4e(t *testing.T) {
	tt := []struct {
		name   string
		config *config.Config
		err    error
	}{
		{
			"single",
			&config.Config{
				EmacsCfgDir: "/home/user/.emacs.d",
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
		generatemu4e(tc.config)
		got, err := os.ReadFile(path.Join(tc.config.EmacsCfgDir, "mu4e.el"))
		if err != nil {
			t.Fatalf("file not saved: %v", err)

		}
		want := fixture("/generatemu4e/" + tc.name + "/mu4e.el")
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("%s: got: %s, want: %s", tc.name, got, want)
		}

	}
}
