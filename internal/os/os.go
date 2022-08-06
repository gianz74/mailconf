package os

import (
	"os"

	"github.com/spf13/afero"
)

var (
	fs = &afero.Afero{
		Fs: afero.NewOsFs(),
	}
	MemFs    = afero.NewMemMapFs()
	OsFs     = afero.NewOsFs()
	Stderr   = os.Stderr
	Stdin    = os.Stdin
	Stdout   = os.Stdout
	ModePerm = os.ModePerm
)

type File = os.File

func Set(f afero.Fs) {
	if f == MemFs {
		f = afero.NewMemMapFs()
		MemFs = f
	}
	fs.Fs = f
}

func MkdirAll(path string, perm os.FileMode) error {
	return fs.MkdirAll(path, perm)
}

func UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func UserConfigDir() (string, error) {
	return os.UserConfigDir()
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
