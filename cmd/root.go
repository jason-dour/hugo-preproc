// cmd/root
//
// Root command for hugo-preproc.

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config - Configuration structure for a single processor.
type Config struct {
	Path    string `mapstructure:"path"`
	Pattern string `mapstructure:"pattern"`
	Command string `mapstructure:"command"`
}

// Configs - Array of processor configs.
type Configs struct {
	Cfgs []Config `mapstructure:"processors,flow"`
}

var (
	// Variable to process the YAML config into.
	configs Configs

	// Used for flags.
	cfgFile string

	// Cobra definition.
	rootCmd = &cobra.Command{
		Use:   "hugo-preproc",
		Short: "A preprocessor for Hugo",
		Long: `Hugo-preproc is a pre-processor for Hugo that allows for
configured processors to be run on the Hugo datafiles.`,
		Run: func(cmd *cobra.Command, args []string) { runProcessors() },
	}
)

// Execute - Executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

// walkMatch - Walk the tree and look for files matching the provided pattern.
func walkMatch(root, pattern string) ([]string, error) {
	// Initialize the match list.
	var matches []string

	// Walk the tree.
	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
				return err
			} else if matched {
				matches = append(matches, path)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

// init - Command initialization routine.
func init() {
	// Initialize Cobra
	cobra.OnInitialize(initConfig)

	// Define the command line flags.
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.hugo-preproc.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			panic(err)
		}

		// Look in HOME and current working directory for the config file.
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")

		// Define the config file name.
		viper.SetConfigName(".hugo-preproc")
	}

	// Read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// Unmarshal the configuration into the config struct.
	err = viper.Unmarshal(&configs)
	if err != nil {
		panic(err)
	}
}

// runProcessors - Iterate through the processors configured in the config file.
func runProcessors() {
	// Loop through each processor...
	for i := 0; i < len(configs.Cfgs); i++ {
		// Walk the tree configured in the processor...retrieving the matched files.
		files, err := walkMatch(configs.Cfgs[i].Path, configs.Cfgs[i].Pattern)
		if err != nil {
			panic(err)
		}
		// Loop through each file matched...
		for j := 0; j < len(files); j++ {
			// Map in the additional functions for the template.
			funcMap := template.FuncMap{
				"replace": strings.Replace,
			}

			// Process the command in the processor as a template.
			tmpl, err := template.New("test").Funcs(funcMap).Parse(configs.Cfgs[i].Command)
			if err != nil {
				panic(err)
			}
			// Convert the template to output string.
			var tmplout bytes.Buffer
			err = tmpl.Execute(&tmplout, files[j])

			// Execute the command and grab the output.
			cmd := exec.Command("sh", "-c", tmplout.String())
			cmd.Stdout = os.Stdout
			err = cmd.Run()
			if err != nil {
				panic(err)
			}
		}
	}

	// All done.
	return
}
