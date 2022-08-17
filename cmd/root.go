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

	"github.com/Masterminds/sprig"
	"github.com/fatih/structs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/imdario/mergo"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GitLogEntry - Individual git log entry and changed files.
type GitLogEntry struct {
	Commit *object.Commit
	Stats  object.FileStats
}

// GitAll - Entire Git log.
type GitAll struct {
	Commits []GitLogEntry
	Head    GitLogEntry
}

// GitLog - Configuration structure for processing git log entries.
type GitLog struct {
	Mode     string `mapstructure:"mode"`
	File     string `mapstructure:"file"`
	Template string `mapstructure:"template"`
}

// Git - Configuration for handling git log entries.
type Git struct {
	Path       string   `mapstructure:"path"`
	Processors []GitLog `mapstructure:"processors"`
}

// Processor - Configuration structure for a single processor.
type Exec struct {
	Path    string `mapstructure:"path"`
	Pattern string `mapstructure:"pattern"`
	Command string `mapstructure:"command"`
}

// Configs - Array of processor configs.
type Configs struct {
	Gits       []Git  `mapstructure:"git,flow"`
	Processors []Exec `mapstructure:"exec,flow"`
}

var (
	// version - the version of the program; injected at compile time.
	version string

	// basename - the basename of the program; injected at compile time.
	basename string

	// Variable to process the YAML config into.
	configs Configs

	// Used for flags.
	cfgFile string

	// Cobra definition.
	rootCmd = &cobra.Command{
		Use:   basename,
		Short: "A preprocessor for Hugo",
		Long: `Hugo-preproc is a pre-processor for Hugo that allows for
configured processors to be run on the Hugo datafiles.`,
		Run:     func(cmd *cobra.Command, args []string) { run() },
		Version: version,
	}

	// Map in the additional functions for the template.
	// TODO: #8 Deprecate custom template functions in favor of the sprig functions.
	funcMap = template.FuncMap{
		"fields":     structs.Names,
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

	// Merge the sprig template functions into our custom template functions.
	err := mergo.Merge(&funcMap, sprig.TxtFuncMap())
	panicIfError(err)
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

// runExec - Iterate through the exec command processors in the config file.
func runExec() {
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

// runGitHead - Process Head mode git log processor.
func runGitHead(repo *git.Repository, ref *plumbing.Reference, processor GitLog) {
	// Grab the HEAD commit.
	commit, err := repo.CommitObject(ref.Hash())
	panicIfError(err)
	commitStats, err := commit.Stats()
	panicIfError(err)

	// Process the file in the config as a template to create the file name.
	fileTemplate, err := template.New("fileTemplate").Funcs(funcMap).Parse(processor.File)
	panicIfError(err)
	var templateFile bytes.Buffer
	err = fileTemplate.Execute(&templateFile, GitLogEntry{
		Commit: commit,
		Stats:  commitStats,
	})
	panicIfError(err)

	// Process the output template in the config.
	outTemplate, err := template.New("outTemplate").Funcs(funcMap).Parse(processor.Template)
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

// runGitEach - Process Each mode git log processor.
func runGitEach(repo *git.Repository, ref *plumbing.Reference, processor GitLog) {
	// Get the commit history in an interator.
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	panicIfError(err)

	// Iterate through the commits.
	err = commitIter.ForEach(func(commit *object.Commit) error {
		// Grab the commit stats.
		commitStats, err := commit.Stats()
		panicIfError(err)
		defer commitIter.Close()

		// Process the file in the config as a template to create the file name.
		fileTemplate, err := template.New("fileTemplate").Funcs(funcMap).Parse(processor.File)
		panicIfError(err)
		var templateFile bytes.Buffer
		err = fileTemplate.Execute(&templateFile, GitLogEntry{
			Commit: commit,
			Stats:  commitStats,
		})
		panicIfError(err)

		// Process the output template in the config.
		outTemplate, err := template.New("outTemplate").Funcs(funcMap).Parse(processor.Template)
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

// runGitAll - Process All mode git log processor.
func runGitAll(repo *git.Repository, ref *plumbing.Reference, processor GitLog) {
	// Get the commit history in an interator.
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	panicIfError(err)
	defer commitIter.Close()

	var allGit GitAll

	// Grab the HEAD commit.
	allGit.Head.Commit, err = repo.CommitObject(ref.Hash())
	panicIfError(err)
	allGit.Head.Stats, err = allGit.Head.Commit.Stats()
	panicIfError(err)

	// Iterate through the commits.
	// allGit.Commits = []GitLogEntry{}
	err = commitIter.ForEach(func(commit *object.Commit) error {
		commitStats, err := commit.Stats()
		panicIfError(err)
		allGit.Commits = append(allGit.Commits, GitLogEntry{
			Commit: commit,
			Stats:  commitStats,
		})
		return nil
	})
	panicIfError(err)

	// Process the file in the config as a template to create the file name.
	fileTemplate, err := template.New("fileTemplate").Funcs(funcMap).Parse(processor.File)
	panicIfError(err)
	var templateFile bytes.Buffer
	err = fileTemplate.Execute(&templateFile, allGit)
	panicIfError(err)

	// Process the output template in the config.
	outTemplate, err := template.New("outTemplate").Funcs(funcMap).Parse(processor.Template)
	panicIfError(err)
	var templateOut bytes.Buffer
	err = outTemplate.Execute(&templateOut, allGit)
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

// runGits - Process the configured git log handlers.
func runGits() {
	// Iterate through the configured git log handlers.
	for i := 0; i < len(configs.Gits); i++ {
		// Define the path to the git repository.
		var repoPath string
		if configs.Gits[i].Path != "" {
			repoPath = configs.Gits[i].Path
		} else {
			repoPath = "."
		}

		// Open the repository.
		repo, err := git.PlainOpen(repoPath)
		panicIfError(err)

		// Get the HEAD commit.
		ref, err := repo.Head()
		panicIfError(err)

		// Iterate through the configured processors.
		for j := 0; j < len(configs.Gits[i].Processors); j++ {
			if configs.Gits[i].Processors[j].Mode != "" {
				// Get the mode.
				switch strings.ToLower(configs.Gits[i].Processors[j].Mode) {
				case "head":
					// Process the HEAD git config.
					runGitHead(repo, ref, configs.Gits[i].Processors[j])
				case "each":
					// Process the Each git config.
					runGitEach(repo, ref, configs.Gits[i].Processors[j])
				case "all":
					// Process the All git config.
					runGitAll(repo, ref, configs.Gits[i].Processors[j])
				}
			}
		}
	}
}

// run - Run the program.
func run() {
	// Run the git processors.
	runGits()

	// Run the file find processors.
	runExec()
}
