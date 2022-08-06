package memcred

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/gianz74/mailconf/internal/cred"
)

type InMemStore map[string]string

func New() InMemStore {
	return make(InMemStore)
}

func (s InMemStore) Add(user, service, host string, port uint16, pwd string) error {
	key := fmt.Sprintf("%s://%s@%s:%d", service, user, host, port)
	fmt.Printf("store: %+v\n", s)
	fmt.Printf("key: [%s]\n", key)
	_, ok := s[key]
	if ok {
		fmt.Printf("returning ErrExistingCreds\n")
		return cred.ErrExistingCreds
	}
	s[key] = pwd
	return nil
}
func (s InMemStore) Delete(user, service, host string, port uint16) error {
	key := fmt.Sprintf("%s://%s@%s:%d", service, user, host, port)
	_, ok := s[key]
	if !ok {
		return cred.ErrNoCreds
	}
	delete(s, key)
	return nil
}
func (s InMemStore) Get(user, service, host string, port uint16) (string, error) {
	key := fmt.Sprintf("%s://%s@%s:%d", service, user, host, port)
	_, ok := s[key]
	if !ok {
		fmt.Printf("no creds for key: %s\n", key)
		return "", cred.ErrNoCreds
	}
	fmt.Printf("creds for key %s: %s\n", key, s[key])
	return s[key], nil
}
func (s InMemStore) Update(user, service, host string, port uint16, pwd string) error {
	key := fmt.Sprintf("%s://%s@%s:%d", service, user, host, port)
	fmt.Printf("updating password for [%s] with: %s\n", key, pwd)
	_, ok := s[key]
	if !ok {
		return cred.ErrNoCreds
	}
	s[key] = pwd
	return nil
}
func (s InMemStore) AddBulk(creds []string) error {
	var service, user, pwd, host string
	var port uint16
	exp := regexp.MustCompile(`^(?P<service>[^:]*)://(?P<user>[^:]*):(?P<pwd>[^@]*)@(?P<host>[^:]*):(?P<port>\d+)$`)
	for k := range s {
		delete(s, k)
	}
	for _, c := range creds {
		match := exp.FindStringSubmatch(c)
		result := make(map[string]string)
		for i, name := range exp.SubexpNames() {
			if i != 0 && name != "" {
				if i < len(match) {
					result[name] = match[i]
				}
			}
		}
		service = result["service"]
		user = result["user"]
		pwd = result["pwd"]
		host = result["host"]
		tmp, err := strconv.ParseUint(result["port"], 10, 16)
		if err != nil {
			return err
		}
		port = uint16(tmp)

		key := fmt.Sprintf("%s://%s@%s:%d", service, user, host, port)
		fmt.Printf("service: %s\n", service)
		fmt.Printf("cred: %s\n", c)
		fmt.Printf("key: %s\n", key)
		_, ok := s[key]
		if ok {
			fmt.Printf("existing creds\n")
			return cred.ErrExistingCreds
		}
		s[key] = pwd
	}
	return nil
}
