package cmd

import (
	"github.com/cqroot/shcompat/pkg/linter"
	"github.com/cqroot/shcompat/pkg/version"
	"github.com/spf13/cobra"
)

var (
	flagVerbose bool
)

func RunRootCmd(cmd *cobra.Command, args []string) {
	var target string
	target = "."
	if len(args) > 0 {
		target = args[0]
	}

	l, err := linter.New(target, flagVerbose)
	cobra.CheckErr(err)
	err = l.Run()
	cobra.CheckErr(err)
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

	c.Version = version.Get().String()

	return &c
}

func Execute() {
	err := NewRootCmd().Execute()
	cobra.CheckErr(err)
}
