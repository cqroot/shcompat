package linter

// WithVerbose sets the verbose option for the linter.
func WithVerbose(verbose bool) Option {
	return func(l *Linter) {
		l.verbose = verbose
	}
}

// WithFormat sets the output format for the linter.
func WithFormat(format Format) Option {
	return func(l *Linter) {
		l.format = format
	}
}

// WithIncludeRules sets the rules to include in the linting process.
func WithIncludeRules(rules []string) Option {
	return func(l *Linter) {
		l.includeRules = rules
	}
}

// WithExcludeRules sets the rules to exclude from the linting process.
func WithExcludeRules(rules []string) Option {
	return func(l *Linter) {
		l.excludeRules = rules
	}
}
