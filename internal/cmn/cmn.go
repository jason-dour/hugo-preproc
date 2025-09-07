// Package cmn implements common variables and utility functions for hugo-preproc,
// providing debug and configuration.
package cmn

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/viper"
)

type (
	GitLogEntry struct {
		Commit *object.Commit
		Stats  object.FileStats
	} // GitLogEntry - Individual git log entry and changed files.

	GitAll struct {
		Commits []GitLogEntry
		Head    GitLogEntry
	} // GitAll - Entire Git log.

	GitProcessor struct {
		Mode     string `mapstructure:"mode"`
		File     string `mapstructure:"file"`
		Template string `mapstructure:"template"`
		Script   string `mapstructure:"script"`
	} // GitProcessor - Configuration structure for processing git log entries.

	Git struct {
		Path       string         `mapstructure:"path"`
		Processors []GitProcessor `mapstructure:"processors"`
	} // Git - Configuration for handling git log entries.

	ExecProcessor struct {
		Path    string `mapstructure:"path"`
		Pattern string `mapstructure:"pattern"`
		Command string `mapstructure:"command"`
		Script  string `mapstructure:"script"`
		Mode    string `mapstructure:"mode"`
	} // ExecProcessor - Configuration structure for a single exec.

	Configs struct {
		Gits  []Git           `mapstructure:"git,flow"`
		Execs []ExecProcessor `mapstructure:"exec,flow"`
	} // Configs - Array of processor configs.
)

var (
	CfgFile string // Used for flags.

	DebugFlag bool // Whether debug output is enabled.
)

var (
	Basename string                // Base name of the program; injected during compile.
	Version  string                // Version of the program; injected during compile.
	Commit   string                // Commit hash of the Version; injected during compile.
	Config   *Configs = &Configs{} // Global configuration for the program.
)

// Debug writes debug output to Stderr if DebugFlag is true.
func Debug(format string, args ...any) {
	if DebugFlag {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("debug: "+format, args...))
	}
}

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {
	funcName := "cmn.InitConfig"
	Debug("%s: begin", funcName)

	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
		Debug("%s: config file set by arg: %s", funcName, CfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("No home directory found; looking for config file only in current directory.\n")
			Debug("%s: no home directory found: %v", funcName, err.Error())
		} else {
			// Look in the home directory for the config file.
			viper.AddConfigPath(home)
			Debug("%s: added home directory to config search path: %s", funcName, home)
		}

		// Look in current working directory for the config file.
		viper.AddConfigPath(".")
		Debug("%s: added current directory to config search path", funcName)

		// Define the config file name.
		viper.SetConfigName(".hugo-preproc")
		Debug("%s: default config file prefix: .hugo-preproc", funcName)
	}

	// Read in environment variables that match
	viper.AutomaticEnv()
	Debug("%s: checking for config environment variables", funcName)

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: error processing config file: %s\n", err.Error())
		os.Exit(1)
	} else {
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}

	// Unmarshal the configuration into the config struct.
	err = viper.Unmarshal(&Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: error unmarshaling config file: %s\n", err.Error())
		os.Exit(1)
	}

	err = checkConfig(Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}

	Debug("%s: end", funcName)
}

// checkConfig checks for configuration cases that conflict.
func checkConfig(configs *Configs) error {
	funcName := "cmn.checkConfig"
	Debug("%s: begin", funcName)

	Debug("%s: checking execs", funcName)
	for i := range configs.Execs {
		Debug("%s: exec %d", funcName, i)
		if (len(configs.Execs[i].Command) > 0) && (len(configs.Execs[i].Script) > 0) {
			Debug("%s: exec %d: config conflict; both command and script defined", funcName, i)
			return fmt.Errorf("%s: exec %d: config conflict; both command and script defined", funcName, i)
		}
	}

	Debug("%s: checking gits", funcName)
	for j := range configs.Gits {
		Debug("%s: git %d: checking processors", funcName, j)
		for k := range configs.Gits[j].Processors {
			Debug("%s: git %d: processor %d", funcName, j, k)
			if (len(configs.Gits[j].Processors[k].Template) > 0) && (len(configs.Gits[j].Processors[k].Script) > 0) {
				Debug("%s: git %d: processor: %d: config conflict; both template and script defined", funcName, j, k)
				return fmt.Errorf("%s: git %d: processor: %d: config conflict; both template and script defined", funcName, j, k)
			}
		}
	}

	Debug("%s: end", funcName)
	return nil
}

// WalkMatch walks the tree and look for files matching the provided pattern.
func WalkMatch(root, pattern string) ([]string, error) {
	funcName := "cmn.WalkMatch"
	Debug("%s: begin", funcName)

	// Initialize the match list.
	var matches []string

	// Walk the tree.
	err := filepath.WalkDir(root,
		func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				Debug("%s: skipping directory: %s", funcName, path)
				return nil
			}
			if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
				return err
			} else if matched {
				Debug("%s: found match: %s", funcName, path)
				matches = append(matches, path)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	Debug("%s: found %d matches", funcName, len(matches))

	Debug("%s: end", funcName)
	return matches, nil
}
