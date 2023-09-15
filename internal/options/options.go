package options

type Option func(*options)

var (
	opts = &options{}
)

type options struct {
	dryrun  bool
	verbose bool
}

func (o *options) parseOptions(opts ...Option) {
	for _, option := range opts {
		option(o)
	}
}

func OptDryrun(dryrun bool) Option {
	return func(o *options) {
		o.dryrun = dryrun
	}
}

func OptVerbose(verbose bool) Option {
	return func(o *options) {
		o.verbose = verbose
	}
}

func Set(o ...Option) {
	opts.parseOptions(o...)
}

func Dryrun() bool {
	return opts.dryrun
}

func Verbose() bool {
	return opts.verbose
}
