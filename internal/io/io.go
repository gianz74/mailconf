package io

import (
	"fmt"
	"io/fs"
	"path"

	"github.com/gianz74/mailconf/internal/options"
	"github.com/gianz74/mailconf/internal/os"
)

type (
	WriteFunc func(string, []byte, fs.FileMode) error
)

func Write(out string, in []byte, perm fs.FileMode) error {
	if options.Dryrun() {
		fmt.Printf("writing to %s\n", out)
	}
	if options.Verbose() {
		fmt.Println(string(in))
	}
	if options.Dryrun() {
		return nil
	}
	err := os.MkdirAll(path.Dir(out), 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(out, in, perm)
	if err != nil {
		return err
	}
	return nil
}
