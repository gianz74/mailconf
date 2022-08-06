package io

import (
	"fmt"
	"io/fs"
	"path"

	"github.com/gianz74/mailconf/internal/os"
)

type WriteFunc func(string, []byte, fs.FileMode) error

var (
	write WriteFunc = GetWriter(false, false)
)

func dryRunWrite(out string, in []byte, verbose bool, perm fs.FileMode) error {
	fmt.Printf("dry-run: writing to %s\n", out)
	if verbose {
		fmt.Println(string(in))
	}
	return nil
}

func realWrite(out string, in []byte, verbose bool, perm fs.FileMode) error {
	err := os.MkdirAll(path.Dir(out), 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(out, in, perm)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Printf("real: writing to %s\n", out)
		fmt.Println(string(in))
	}
	return nil
}

func SetWriter(w WriteFunc) WriteFunc {
	ret := write
	write = w
	return ret
}

func GetWriter(dryrun, verbose bool) WriteFunc {
	return func(out string, in []byte, perm fs.FileMode) error {
		if dryrun {
			return dryRunWrite(out, in, verbose, perm)
		}
		return realWrite(out, in, verbose, perm)
	}
}

func Write(out string, in []byte, perm fs.FileMode) error {
	return write(out, in, perm)
}
