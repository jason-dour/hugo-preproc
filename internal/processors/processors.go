// Package processors provide the various functions to run processors.
package processors

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/jason-dour/hugo-preproc/internal/cmn"
)

// Execs iterates through the exec command processors in the config file.
func Execs(configs *cmn.Configs) error {
	// Loop through each processor...
	for i := range configs.Processors {
		// Walk the tree configured in the processor...retrieving the matched files.
		files, err := cmn.WalkMatch(configs.Processors[i].Path, configs.Processors[i].Pattern)
		if err != nil {
			return err
		}

		// Loop through each file matched...
		for j := range files {
			// Process the command in the processor as a template.
			outTemplate, err := template.New("outTemplate").Funcs(cmn.FuncMap).Parse(configs.Processors[i].Command)
			if err != nil {
				return err
			}

			// Convert the template to output string.
			var templateOut bytes.Buffer
			err = outTemplate.Execute(&templateOut, files[j])
			if err != nil {
				return err
			}

			// Execute the command and grab the output.
			cmd := exec.Command("sh", "-c", templateOut.String())
			cmd.Stdout = os.Stdout
			err = cmd.Run()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// gitHead - Process Head mode git log processor.
func gitHead(repo *git.Repository, ref *plumbing.Reference, processor cmn.GitLog) error {
	// Grab the HEAD commit.
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return err
	}
	commitStats, err := commit.Stats()
	if err != nil {
		return err
	}

	// Process the file in the config as a template to create the file name.
	fileTemplate, err := template.New("fileTemplate").Funcs(cmn.FuncMap).Parse(processor.File)
	if err != nil {
		return err
	}
	var templateFile bytes.Buffer
	err = fileTemplate.Execute(&templateFile, cmn.GitLogEntry{
		Commit: commit,
		Stats:  commitStats,
	})
	if err != nil {
		return err
	}

	// Process the output template in the config.
	outTemplate, err := template.New("outTemplate").Funcs(cmn.FuncMap).Parse(processor.Template)
	if err != nil {
		return err
	}
	var templateOut bytes.Buffer
	err = outTemplate.Execute(&templateOut, cmn.GitLogEntry{
		Commit: commit,
		Stats:  commitStats,
	})
	if err != nil {
		return err
	}

	// Create the file.
	err = os.MkdirAll(filepath.Dir(templateFile.String()), 0755)
	if err != nil {
		return err
	}
	outFile, err := os.Create(templateFile.String())
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Write the output to the file.
	_, err = outFile.WriteString(templateOut.String())
	if err != nil {
		return err
	}

	return nil
}

// gitEach - Process Each mode git log processor.
func gitEach(repo *git.Repository, ref *plumbing.Reference, processor cmn.GitLog) error {
	// Get the commit history in an interator.
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return err
	}

	// Iterate through the commits.
	err = commitIter.ForEach(func(commit *object.Commit) error {
		// Grab the commit stats.
		commitStats, err := commit.Stats()
		if err != nil {
			return err
		}
		defer commitIter.Close()

		// Process the file in the config as a template to create the file name.
		fileTemplate, err := template.New("fileTemplate").Funcs(cmn.FuncMap).Parse(processor.File)
		if err != nil {
			return err
		}
		var templateFile bytes.Buffer
		err = fileTemplate.Execute(&templateFile, cmn.GitLogEntry{
			Commit: commit,
			Stats:  commitStats,
		})
		if err != nil {
			return err
		}

		// Process the output template in the config.
		outTemplate, err := template.New("outTemplate").Funcs(cmn.FuncMap).Parse(processor.Template)
		if err != nil {
			return err
		}
		var templateOut bytes.Buffer
		err = outTemplate.Execute(&templateOut, cmn.GitLogEntry{
			Commit: commit,
			Stats:  commitStats,
		})
		if err != nil {
			return err
		}

		// Create the file.
		err = os.MkdirAll(filepath.Dir(templateFile.String()), 0755)
		if err != nil {
			return err
		}
		outFile, err := os.Create(templateFile.String())
		if err != nil {
			return err
		}
		defer outFile.Close()

		// Write the output to the file.
		_, err = outFile.WriteString(templateOut.String())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// gitAll - Process All mode git log processor.
func gitAll(repo *git.Repository, ref *plumbing.Reference, processor cmn.GitLog) error {
	// Get the commit history in an interator.
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return err
	}
	defer commitIter.Close()

	var allGit cmn.GitAll

	// Grab the HEAD commit.
	allGit.Head.Commit, err = repo.CommitObject(ref.Hash())
	if err != nil {
		return err
	}
	allGit.Head.Stats, err = allGit.Head.Commit.Stats()
	if err != nil {
		return err
	}

	// Iterate through the commits.
	// allGit.Commits = []GitLogEntry{}
	err = commitIter.ForEach(func(commit *object.Commit) error {
		commitStats, err := commit.Stats()
		if err != nil {
			return err
		}
		allGit.Commits = append(allGit.Commits, cmn.GitLogEntry{
			Commit: commit,
			Stats:  commitStats,
		})
		return nil
	})
	if err != nil {
		return err
	}

	// Process the file in the config as a template to create the file name.
	fileTemplate, err := template.New("fileTemplate").Funcs(cmn.FuncMap).Parse(processor.File)
	if err != nil {
		return err
	}
	var templateFile bytes.Buffer
	err = fileTemplate.Execute(&templateFile, allGit)
	if err != nil {
		return err
	}

	// Process the output template in the config.
	outTemplate, err := template.New("outTemplate").Funcs(cmn.FuncMap).Parse(processor.Template)
	if err != nil {
		return err
	}
	var templateOut bytes.Buffer
	err = outTemplate.Execute(&templateOut, allGit)
	if err != nil {
		return err
	}

	// Create the file.
	err = os.MkdirAll(filepath.Dir(templateFile.String()), 0755)
	if err != nil {
		return err
	}
	outFile, err := os.Create(templateFile.String())
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Write the output to the file.
	_, err = outFile.WriteString(templateOut.String())
	if err != nil {
		return err
	}

	return nil
}

// Gits - Process the configured git log handlers.
func Gits(configs *cmn.Configs) error {
	// Iterate through the configured git log handlers.
	for i := range configs.Gits {
		// Define the path to the git repository.
		var repoPath string
		if configs.Gits[i].Path != "" {
			repoPath = configs.Gits[i].Path
		} else {
			repoPath = "."
		}

		// Open the repository.
		repo, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{
			DetectDotGit:          true,
			EnableDotGitCommonDir: true,
		})
		if err != nil {
			return err
		}

		// Get the HEAD commit.
		ref, err := repo.Head()
		if err != nil {
			return err
		}

		// Iterate through the configured processors.
		for j := range configs.Gits[i].Processors {
			if configs.Gits[i].Processors[j].Mode != "" {
				// Get the mode.
				switch strings.ToLower(configs.Gits[i].Processors[j].Mode) {
				case "head":
					// Process the HEAD git config.
					err := gitHead(repo, ref, configs.Gits[i].Processors[j])
					if err != nil {
						return err
					}
				case "each":
					// Process the Each git config.
					err := gitEach(repo, ref, configs.Gits[i].Processors[j])
					if err != nil {
						return err
					}
				case "all":
					// Process the All git config.
					err := gitAll(repo, ref, configs.Gits[i].Processors[j])
					if err != nil {
						return err
					}
				default:
					err := fmt.Errorf("invalid git processor mode; should be head/each/all")
					return err
				}
			}
		}
	}

	return nil
}
