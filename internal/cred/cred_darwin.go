//go:build darwin

package cred

import (
	"fmt"

	"github.com/gianz74/mailconf/internal/os"
)

var (
	Credentials CredentialsStore = Darwin{}
)

type CredentialsStore interface {
	Add(user, service, host string, port uint16, pwd string) error
	Get(user, service, host string, port uint16) (string, error)
	Delete(user, service, host string, port uint16) error
	Update(user, service, host string, port uint16, pwd string) error
}

func init() {
	if os.System == "darwin" {
		Credentials = Darwin{}
	}
}

type Darwin map[string]string

func (Darwin) Add(user, service, host string, port uint16, pwd string) error {
	fmt.Fprintf(os.Stderr, "would add credentials for %s:%s@%s://%s:%d\n", user, pwd, service, host, port)
	return nil
}

func (Darwin) Get(user, service, host string, port uint16) (string, error) {
	return "", nil
}

func (Darwin) Delete(user, service, host string, port uint16) error {
	return nil
}

func (Darwin) Update(user, service, host string, port uint16, pwd string) error {
	return nil
}
