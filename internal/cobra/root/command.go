// Package root provides root command processing for the command.
package root

import (
	"github.com/jason-dour/hugo-preproc/internal/cmn"
	"github.com/jason-dour/hugo-preproc/internal/processors"
	"github.com/spf13/cobra"
)

var (
	Cmd = &cobra.Command{
		Use:     cmn.Basename,
		Short:   "A preprocessor for Hugo",
		Long:    "hugo-preproc is a pre-processor for Hugo that allows for configured\nprocessors to be run on the Hugo datafiles.",
		RunE:    run,
		Version: cmn.Version + " (" + cmn.Commit + ")",
	} // Root command definition.
)

// init - Command initialization routine.
func init() {
	// Initialize Cobra
	cobra.OnInitialize(cmn.InitConfig)

	// Command flags.
	Cmd.PersistentFlags().StringVarP(&cmn.CfgFile, "config", "c", "", "config file (default is $HOME/.hugo-preproc.yaml)")
	Cmd.PersistentFlags().BoolVarP(&cmn.DebugFlag, "debug", "d", false, "enable debug mode")
}

// run - Run the program.
func run(cmd *cobra.Command, args []string) error {
	cmn.Debug("run: begin")

	// Run the git processors.
	cmn.Debug("run: running git processors")
	err := processors.Gits(cmn.Config)
	if err != nil {
		return err
	}

	// Run the file find processors.
	cmn.Debug("run: running exec processors")
	err = processors.Execs(cmn.Config)
	if err != nil {
		return err
	}

	cmn.Debug("run: end")
	return nil
}
