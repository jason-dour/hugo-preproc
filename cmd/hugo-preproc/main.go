// hugo-preproc is a pre-processor for Hugo that allows for
// configured processors to be run on the Hugo datafiles.
//
// Usage:
//
//	hugo-preproc [flags]
//
// Flags:
//
//	-c, --config string   config file (default is $HOME/.hugo-preproc.yaml)
//	-d, --debug           enable debug mode
//	-h, --help            help for hugo-preproc
//	-v, --version         version for hugo-preproc
package main

import (
	"os"

	"github.com/jason-dour/hugo-preproc/internal/cobra/root"
)

func main() {
	err := root.Cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
