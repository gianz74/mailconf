package testutil

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/gianz74/mailconf/internal/cred"
	myos "github.com/gianz74/mailconf/internal/os"
	"github.com/spf13/afero"
)

const (
	prefix = "testdata/fixtures"
)

type Option func(*options)

type options struct {
	name     *string
	subname  *string
	system   *string
	savedata bool
	copy     bool
}

func Name(name string) Option {
	return func(o *options) {
		o.name = &name
	}
}

func SubName(name string) Option {
	return func(o *options) {
		o.subname = &name
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

func CopyFs() Option {
	return func(o *options) {
		o.copy = true
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
	if opts.subname != nil {
		fspath = path.Join(fspath, *opts.subname)
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

	if opts.copy {
		tmpsfs := afero.Afero{
			Fs: base,
		}
		type exch struct {
			filename string
			info     os.FileInfo
		}
		out := make(chan exch)
		go func(o chan exch) {
			rootpath := "/"
			err := tmpsfs.Walk(rootpath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					o <- exch{path[len(rootpath):], info}
				}
				return nil
			})
			close(o)
			if err != nil {
				panic(err)
			}
		}(out)
		ret := afero.Afero{
			Fs: shadow,
		}
		for ex := range out {
			src, _ := tmpsfs.ReadFile(ex.filename)
			ret.WriteFile(ex.filename, src, ex.info.Mode())
		}
		return ret
	}
	ret := afero.Afero{}

	ret.Fs = afero.NewCopyOnWriteFs(base, shadow)

	return ret
}

func Fixture(name, subname, system, file string) string {
	b, err := ioutil.ReadFile(path.Join(prefix, name, subname, system, "want", file))
	if err != nil {
		panic(err)
	}
	return string(b)
}

func Result(file string) (string, error) {
	b, err := myos.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func GetFiles(name, subname, system string) <-chan string {
	out := make(chan string)
	go func(o chan string) {
		rootpath := path.Join(prefix, name, subname, system, "want")
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

func CheckCreds(creds string) (string, string) {
	exp := regexp.MustCompile(`^(?P<service>[^:]*)://(?P<user>[^:]*):(?P<pwd>[^@]*)@(?P<host>[^:]*):(?P<port>\d+)$`)
	match := exp.FindStringSubmatch(creds)
	result := make(map[string]string)
	for i, name := range exp.SubexpNames() {
		if i != 0 && name != "" {
			if i < len(match) {
				result[name] = match[i]
			}
		}
	}
	service, ok := result["service"]
	if !ok {
		panic("no creds")
	}
	user, ok := result["user"]
	if !ok {
		panic("no creds")
	}
	pwd, ok := result["pwd"]
	if !ok {
		panic("no creds")
	}
	host, ok := result["host"]
	if !ok {
		panic("no creds")
	}
	port_s, ok := result["port"]
	if !ok {
		panic("no creds")
	}

	tmp, _ := strconv.ParseUint(port_s, 10, 16)
	port := uint16(tmp)
	credstore := cred.New()
	got, _ := credstore.Get(user, service, host, port)
	return got, pwd
}
