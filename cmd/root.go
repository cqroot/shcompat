package cmd

import (
	"fmt"
	"os"

	"github.com/cqroot/shcompat/pkg/linter"
	"github.com/cqroot/shcompat/pkg/version"
	"github.com/spf13/cobra"
)

var (
	flagVerbose      bool
	flagFormat       string
	flagIncludeRules []string
	flagExcludeRules []string
)

func CheckErr(msg interface{}) {
	if msg == nil {
		return
	}

	fmt.Fprintln(os.Stderr, "Error:", msg)
	os.Exit(int(linter.ResultError))
}

func RunRootCmd(cmd *cobra.Command, args []string) {
	var target string
	target = "."
	if len(args) > 0 {
		target = args[0]
	}

	outputFormat, err := linter.ParseFormat(flagFormat)
	CheckErr(err)

	l, err := linter.New(target,
		linter.WithVerbose(flagVerbose),
		linter.WithFormat(outputFormat),
		linter.WithIncludeRules(flagIncludeRules),
		linter.WithExcludeRules(flagExcludeRules),
	)
	CheckErr(err)
	ret, err := l.Run()
	CheckErr(err)
	os.Exit(int(ret))
}

func NewRootCmd() *cobra.Command {
	c := cobra.Command{
		Use:   "shcompat",
		Short: "A Shell Compatibility Linter.",
		Long:  "A Shell Compatibility Linter.",
		Args:  cobra.MatchAll(cobra.RangeArgs(0, 1), cobra.OnlyValidArgs),
		Run:   RunRootCmd,
	}
	c.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "verbose output")
	c.PersistentFlags().StringVarP(&flagFormat, "format", "f", "tty", "output format: tty, json")
	c.PersistentFlags().StringSliceVarP(&flagIncludeRules, "include-rules", "i", nil, "include only the specified rules")
	c.PersistentFlags().StringSliceVarP(&flagExcludeRules, "exclude-rules", "e", nil, "exclude the specified rules")

	c.Version = version.Get().String()

	return &c
}

func Execute() {
	err := NewRootCmd().Execute()
	CheckErr(err)
}
