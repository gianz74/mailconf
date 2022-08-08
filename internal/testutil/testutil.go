package testutil

import (
	"path"

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
