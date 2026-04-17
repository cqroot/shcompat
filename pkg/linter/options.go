package linter

func WithVerbose(verbose bool) Option {
	return func(l *Linter) {
		l.verbose = verbose
	}
}

func WithFormat(format Format) Option {
	return func(l *Linter) {
		l.format = format
	}
}
