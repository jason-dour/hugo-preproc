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

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type GitLogEntry struct {
	Commit *object.Commit
	Stats  object.FileStats
}

// GitLog - Configuration structure for processing git log entries.
type GitLog struct {
	File     string `mapstructure:"file"`
	Template string `mapstructure:"template"`
}

// Git - Configuration for handling git log entries.
type Git struct {
	Path string `mapstructure:"path"`
	Head GitLog `mapstructure:"head"`
	Each GitLog `mapstructure:"each"`
	All  GitLog `mapstructure:"all"`
}

// Processor - Configuration structure for a single processor.
type Processor struct {
	Path    string `mapstructure:"path"`
	Pattern string `mapstructure:"pattern"`
	Command string `mapstructure:"command"`
}

// Configs - Array of processor configs.
type Configs struct {
	Gits       Git         `mapstructure:"git"`
	Processors []Processor `mapstructure:"processors,flow"`
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
		Run: func(cmd *cobra.Command, args []string) { run() },
	}

	// Map in the additional functions for the template.
	funcMap = template.FuncMap{
		"replace":    strings.Replace,
		"split":      strings.Split,
		"trimsuffix": strings.TrimSuffix,
	}
)

// panicIfError - Panic if an error occurred.
func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

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
			panicIfError(err)

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
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
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
	for i := 0; i < len(configs.Processors); i++ {
		// Walk the tree configured in the processor...retrieving the matched files.
		files, err := walkMatch(configs.Processors[i].Path, configs.Processors[i].Pattern)
		panicIfError(err)

		// Loop through each file matched...
		for j := 0; j < len(files); j++ {
			// Process the command in the processor as a template.
			outTemplate, err := template.New("outTemplate").Funcs(funcMap).Parse(configs.Processors[i].Command)
			panicIfError(err)

			// Convert the template to output string.
			var templateOut bytes.Buffer
			err = outTemplate.Execute(&templateOut, files[j])
			panicIfError(err)

			// Execute the command and grab the output.
			cmd := exec.Command("sh", "-c", templateOut.String())
			cmd.Stdout = os.Stdout
			err = cmd.Run()
			panicIfError(err)
		}
	}
}

func runGits() {
	// Define the path to the git repository.
	var repoPath string
	if configs.Gits.Path != "" {
		repoPath = configs.Gits.Path
	} else {
		repoPath = "."
	}

	// Open the repository.
	r, err := git.PlainOpen(repoPath)
	panicIfError(err)

	// Get the HEAD commit.
	ref, err := r.Head()
	panicIfError(err)

	// Get the commit history in an interator.
	commitIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	panicIfError(err)
	_ = commitIter

	// Handle the Head git config if it exists.
	if (len(configs.Gits.Head.File) > 0) && (len(configs.Gits.Head.Template) > 0) {
		// Grab the HEAD commit.
		commit, err := r.CommitObject(ref.Hash())
		panicIfError(err)
		commitStats, err := commit.Stats()
		panicIfError(err)

		// Process the file in the config as a template to create the file name.
		fileTemplate, err := template.New("fileTemplate").Funcs(funcMap).Parse(configs.Gits.Head.File)
		panicIfError(err)
		var templateFile bytes.Buffer
		err = fileTemplate.Execute(&templateFile, GitLogEntry{
			Commit: commit,
			Stats:  commitStats,
		})
		panicIfError(err)

		// Process the output template in the config.
		outTemplate, err := template.New("outTemplate").Funcs(funcMap).Parse(configs.Gits.Head.Template)
		panicIfError(err)
		var templateOut bytes.Buffer
		err = outTemplate.Execute(&templateOut, GitLogEntry{
			Commit: commit,
			Stats:  commitStats,
		})
		panicIfError(err)

		// Create the file.
		err = os.MkdirAll(filepath.Dir(templateFile.String()), 0755)
		panicIfError(err)
		outFile, err := os.Create(templateFile.String())
		panicIfError(err)
		defer outFile.Close()

		// Write the output to the file.
		_, err = outFile.WriteString(templateOut.String())
		panicIfError(err)
	}

	// Handle the Each git config if it exists.
	if (len(configs.Gits.Each.File) > 0) && (len(configs.Gits.Each.Template) > 0) {
		// Iterate through the commits.
		err = commitIter.ForEach(func(commit *object.Commit) error {
			commitStats, err := commit.Stats()
			panicIfError(err)

			// Process the file in the config as a template to create the file name.
			fileTemplate, err := template.New("fileTemplate").Funcs(funcMap).Parse(configs.Gits.Each.File)
			panicIfError(err)
			var templateFile bytes.Buffer
			err = fileTemplate.Execute(&templateFile, GitLogEntry{
				Commit: commit,
				Stats:  commitStats,
			})
			panicIfError(err)

			// Process the output template in the config.
			outTemplate, err := template.New("outTemplate").Funcs(funcMap).Parse(configs.Gits.Each.Template)
			panicIfError(err)
			var templateOut bytes.Buffer
			err = outTemplate.Execute(&templateOut, GitLogEntry{
				Commit: commit,
				Stats:  commitStats,
			})
			panicIfError(err)

			// Create the file.
			err = os.MkdirAll(filepath.Dir(templateFile.String()), 0755)
			panicIfError(err)
			outFile, err := os.Create(templateFile.String())
			panicIfError(err)
			defer outFile.Close()

			// Write the output to the file.
			_, err = outFile.WriteString(templateOut.String())
			panicIfError(err)

			return nil
		})
		panicIfError(err)
	}

	// TODO: #3 Handle the All git config if it exists.
}

func run() {
	// Run the git log processors.
	runGits()

	// Run the processors.
	runProcessors()
}
