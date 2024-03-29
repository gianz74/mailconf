package os

import (
	"os"
	"runtime"
)

type File = os.File

type FsAccess interface {
	MkdirAll(string, os.FileMode) error
	WriteFile(string, []byte, os.FileMode) error
	ReadFile(string) ([]byte, error)
	RemoveAll(string) error
}

var (
	fs            FsAccess = osFs{}
	Stderr                 = os.Stderr
	Stdin                  = os.Stdin
	Stdout                 = os.Stdout
	ModePerm               = os.ModePerm
	UserConfigDir          = os.UserConfigDir
	UserHomeDir            = os.UserHomeDir
	System                 = runtime.GOOS
)

func Set(f FsAccess) FsAccess {
	ret := fs
	fs = f
	return ret
}

type osFs struct{}

func (osFs) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (osFs) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (osFs) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (osFs) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func MkdirAll(path string, perm os.FileMode) error {
	return fs.MkdirAll(path, perm)
}

func RemoveAll(path string) error {
	return fs.RemoveAll(path)
}

func WriteFile(filename string, data []byte, perm os.FileMode) error {
	return fs.WriteFile(filename, data, perm)
}

func ReadFile(filename string) ([]byte, error) {
	return fs.ReadFile(filename)
}

func Exit(code int) {
	os.Exit(code)
}
