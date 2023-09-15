//go:build linux

package cred

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gianz74/mailconf/internal/options"
	"github.com/gianz74/mailconf/internal/os"
)

var (
	_credentials     CredentialsStore
	ErrExistingCreds = errors.New("Existing credentials")
	ErrNoCreds       = errors.New("Credentials not found")
)

type CredentialsStore interface {
	Add(user, service, host string, port uint16, pwd string) error
	Get(user, service, host string, port uint16) (string, error)
	Delete(user, service, host string, port uint16) error
	Update(user, service, host string, port uint16, pwd string) error
}

func SetStore(s CredentialsStore) CredentialsStore {
	ret := _credentials
	_credentials = s
	return ret
}

func New() CredentialsStore {
	if _credentials == nil {
		_credentials = make(Linux)
	}
	return _credentials
}

type Linux map[string]string

func (c Linux) Add(user, service, host string, port uint16, pwd string) error {
	_, err := c.Get(user, service, host, port)
	if err == nil {
		return ErrExistingCreds
	}

	label := fmt.Sprintf("%s %s password for %s:%d", user, service, host, port)
	if options.Dryrun() {
		if options.Verbose() {
			fmt.Fprintf(os.Stdout, "setting password for %s://%s@%s:%d to %s\n", service, user, host, port, pwd)
		} else {
			fmt.Fprintf(os.Stdout, "setting password for %s://%s@%s:%d\n", service, user, host, port)
		}
		return nil
	}
	cmd := exec.Command("secret-tool", "store", "--label", label, "host", host, "user", user, "port", fmt.Sprintf("%d", port), "service", service)
	cmd.Stdin = strings.NewReader(pwd)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (c Linux) Get(user, service, host string, port uint16) (string, error) {
	cmd := exec.Command("secret-tool", "lookup", "user", user, "host", host, "port", fmt.Sprintf("%d", port), "service", service)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", ErrNoCreds
	}
	return out.String(), nil
}

func (c Linux) Delete(user, service, host string, port uint16) error {
	_, err := c.Get(user, service, host, port)
	if err != nil {
		return err
	}

	if options.Dryrun() {
		fmt.Fprintf(os.Stdout, "removing password for %s://%s@%s:%d\n", service, user, host, port)
		return nil
	}
	cmd := exec.Command("secret-tool", "clear", "user", user, "host", host, "port", fmt.Sprintf("%d", port), "service", service)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (c Linux) Update(user, service, host string, port uint16, pwd string) error {
	_, err := c.Get(user, service, host, port)
	if err != nil {
		return err
	}
	label := fmt.Sprintf("%s %s password for %s:%d", user, service, host, port)
	if options.Dryrun() {
		if options.Verbose() {
			fmt.Fprintf(os.Stdout, "updating password for %s://%s@%s:%d to %s\n", service, user, host, port, pwd)
		} else {
			fmt.Fprintf(os.Stdout, "updating password for %s://%s@%s:%d\n", service, user, host, port)
		}
		return nil
	}
	cmd := exec.Command("secret-tool", "store", "--label", label, "host", host, "user", user, "port", fmt.Sprintf("%d", port), "service", service)
	cmd.Stdin = strings.NewReader(pwd)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
