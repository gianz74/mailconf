package testutil

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	myos "github.com/gianz74/mailconf/internal/os"
	"github.com/spf13/afero"
)

const (
	prefix = "testdata/fixtures"
)

type Option func(*options)

type options struct {
	name     *string
	system   *string
	savedata bool
}

func Name(name string) Option {
	return func(o *options) {
		o.name = &name
	}
}

func System(system string) Option {
	return func(o *options) {
		o.system = &system
	}
}

func SaveData() Option {
	return func(o *options) {
		o.savedata = true
	}
}

func (o *options) parseOptions(opts ...Option) {
	for _, option := range opts {
		option(o)
	}
}

func NewFs(o ...Option) afero.Afero {
	opts := &options{}

	opts.parseOptions(o...)

	fspath := prefix
	if opts.name != nil {
		fspath = path.Join(fspath, *opts.name)
	}
	if opts.system != nil {
		fspath = path.Join(fspath, *opts.system)
	}
	ropath := path.Join(fspath, "root")
	rwpath := path.Join(fspath, "want")
	base := afero.NewBasePathFs(afero.NewReadOnlyFs(afero.NewOsFs()), ropath)
	var shadow afero.Fs
	if opts.savedata {
		shadow = afero.NewBasePathFs(afero.NewOsFs(), rwpath)
	} else {
		shadow = afero.NewMemMapFs()
	}
	ret := afero.Afero{}

	ret.Fs = afero.NewCopyOnWriteFs(base, shadow)

	return ret
}

func Fixture(name, system, file string) string {
	b, err := ioutil.ReadFile(path.Join(prefix, name, system, "want", file))
	if err != nil {
		panic(err)
	}
	return string(b)
}

func Result(file string) string {
	b, err := myos.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func GetFiles(name, system string) <-chan string {
	out := make(chan string)
	go func(o chan string) {
		rootpath := path.Join(prefix, name, system, "want")
		err := filepath.Walk(rootpath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				o <- path[len(rootpath):]
			}
			return nil
		})
		close(o)
		if err != nil {
			panic(err)
		}

	}(out)
	return out
}
