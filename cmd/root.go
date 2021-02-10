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
	configs Configs // Variable to process the YAML config into.

	// Used for flags.
	cfgFile     string
	userLicense string

	rootCmd = &cobra.Command{
		Use:   "hugo-preproc",
		Short: "A preprocessor for Hugo",
		Long: `Hugo-preproc is a pre-processor for Hugo that allows for
configured processors to be run on the Hugo datafiles.`,
		Run: func(cmd *cobra.Command, args []string) { runProcessors() },
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

// walkMatch - TODO
func walkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.hugo-preproc.yaml)")
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
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".hugo-preproc")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	err := viper.Unmarshal(&configs)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("%#v\n", configs)
}

func runProcessors() {
	for i := 0; i < len(configs.Cfgs); i++ {
		files, err := walkMatch(configs.Cfgs[i].Path, configs.Cfgs[i].Pattern)
		if err != nil {
			panic(err)
		}
		for j := 0; j < len(files); j++ {
			funcMap := template.FuncMap{
				"replace": strings.Replace,
			}

			tmpl, err := template.New("test").Funcs(funcMap).Parse(configs.Cfgs[i].Command)
			if err != nil {
				panic(err)
			}
			var tmplout bytes.Buffer
			err = tmpl.Execute(&tmplout, files[j])
			// fmt.Printf("%v\n", tmplout.String())

			cmd := exec.Command("sh", "-c", tmplout.String())
			cmd.Stdout = os.Stdout
			err = cmd.Run()
			if err != nil {
				panic(err)
			}
		}
		// fmt.Printf("%v\n", files)
	}

	return
}
