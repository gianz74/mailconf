package cred

import (
	"strconv"
	"strings"
	"testing"
)

func setup(creds []string) {
	store := New()
	for _, c := range creds {
		fields := strings.FieldsFunc(c, func(c rune) bool { return c == '|' })
		port, _ := strconv.ParseUint(fields[3], 10, 16)

		store.Add(fields[0], fields[1], fields[2], uint16(port), fields[4])
	}
}

func cleanup(creds []string) {
	store := New()
	for _, c := range creds {
		fields := strings.FieldsFunc(c, func(c rune) bool { return c == '|' })
		port, _ := strconv.ParseUint(fields[3], 10, 16)

		store.Delete(fields[0], fields[1], fields[2], uint16(port))
	}
}

func TestAdd(t *testing.T) {
	tt := []struct {
		name    string
		creds   []string
		service string
		host    string
		port    uint16
		user    string
		pass    string
		want    string
		err     error
	}{
		{
			"simple add",
			[]string{},
			"imap",
			"imap.gmail.com",
			993,
			"user@gmail.com",
			"secret",
			"secret",
			nil,
		},
		{
			"add existing",
			[]string{
				"user@gmail.com|imap|imap.gmail.com|993|secret",
			},
			"imap",
			"imap.gmail.com",
			993,
			"user@gmail.com",
			"secret",
			"",
			ErrExistingCreds,
		},
	}
	for _, tc := range tt {
		setup(tc.creds)
		store := New()
		defer cleanup(tc.creds)
		err := store.Add(tc.user, tc.service, tc.host, tc.port, tc.pass)
		if err != tc.err {
			t.Fatalf("%s: got error %v, want: %v", tc.name, err, tc.err)
		}
		if tc.err == nil {
			got, err := store.Get(tc.user, tc.service, tc.host, tc.port)
			if err != nil {
				t.Fatalf("%s: got error %v retrieving expected credentials\n", tc.name, err)
			}
			defer store.Delete(tc.user, tc.service, tc.host, tc.port)
			if got != tc.want {
				t.Fatalf("%s: got \"%s\", want: \"%s\"\n", tc.name, got, tc.want)
			}
		}

	}
}

func TestGet(t *testing.T) {
	tt := []struct {
		name    string
		creds   []string
		service string
		host    string
		port    uint16
		user    string
		want    string
		err     error
	}{
		{
			"get non existent",
			[]string{},
			"imap",
			"imap.gmail.com",
			993,
			"user@gmail.com",
			"",
			ErrNoCreds,
		},
		{
			"get existent",
			[]string{
				"user@gmail.com|imap|imap.gmail.com|993|secret",
			},
			"imap",
			"imap.gmail.com",
			993,
			"user@gmail.com",
			"secret",
			nil,
		},
	}
	for _, tc := range tt {
		setup(tc.creds)
		store := New()
		defer cleanup(tc.creds)
		got, err := store.Get(tc.user, tc.service, tc.host, tc.port)
		if err != tc.err {
			t.Fatalf("%s: got error %v, want: %v", tc.name, err, tc.err)
		}
		if tc.err == nil {
			if got != tc.want {
				t.Fatalf("%s: got \"%s\", want: \"%s\"\n", tc.name, got, tc.want)
			}
		}

	}
}

func TestUpdate(t *testing.T) {
	tt := []struct {
		name    string
		creds   []string
		service string
		host    string
		port    uint16
		user    string
		pass    string
		want    string
		err     error
	}{
		{
			"update existing",
			[]string{
				"user@gmail.com|imap|imap.gmail.com|993|secret",
			},
			"imap",
			"imap.gmail.com",
			993,
			"user@gmail.com",
			"newsecret",
			"newsecret",
			nil,
		},
		{
			"update non existent",
			[]string{
				"user@gmail.com|imap|imap.gmail.com|993|secret",
			},
			"imap",
			"imap.gmail.com",
			993,
			"another@gmail.com",
			"newsecret",
			"",
			ErrNoCreds,
		},
	}
	for _, tc := range tt {
		setup(tc.creds)
		store := New()
		defer cleanup(tc.creds)
		err := store.Update(tc.user, tc.service, tc.host, tc.port, tc.pass)
		if err != tc.err {
			t.Fatalf("%s: got error %v, want: %v", tc.name, err, tc.err)
		}
		if tc.err == nil {
			got, err := store.Get(tc.user, tc.service, tc.host, tc.port)
			if err != nil {
				t.Fatalf("%s: got error %v retrieving expected credentials\n", tc.name, err)
			}
			defer store.Delete(tc.user, tc.service, tc.host, tc.port)
			if got != tc.want {
				t.Fatalf("%s: got \"%s\", want: \"%s\"\n", tc.name, got, tc.want)
			}
		}

	}
}

func TestDelete(t *testing.T) {
	tt := []struct {
		name    string
		creds   []string
		service string
		host    string
		port    uint16
		user    string
		err     error
	}{
		{
			"delete existent",
			[]string{
				"user@gmail.com|imap|imap.gmail.com|993|secret",
			},
			"imap",
			"imap.gmail.com",
			993,
			"user@gmail.com",
			nil,
		},
		{
			"delete non existent",
			[]string{},
			"imap",
			"imap.gmail.com",
			993,
			"user@gmail.com",
			ErrNoCreds,
		},
	}
	for _, tc := range tt {
		setup(tc.creds)
		store := New()
		defer cleanup(tc.creds)
		err := store.Delete(tc.user, tc.service, tc.host, tc.port)
		if err != tc.err {
			t.Fatalf("%s: got error %v, want: %v", tc.name, err, tc.err)
		}
		if tc.err == nil {
			_, err := store.Get(tc.user, tc.service, tc.host, tc.port)
			if err != ErrNoCreds {
				t.Fatalf("%s: got error %v retrieving expected credentials\n", tc.name, err)
			}
		}

	}
}
