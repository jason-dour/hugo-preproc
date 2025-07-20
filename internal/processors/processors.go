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

	"github.com/Masterminds/sprig/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/jason-dour/hugo-preproc/internal/cmn"
)

// Execs iterates through the exec command processors in the config file.
func Execs(configs *cmn.Configs) error {
	funcName := "processors.Execs"
	cmn.Debug("%s: begin", funcName)

	// Loop through each processor...
	cmn.Debug("%s: iterating execs: %d", funcName, len(configs.Execs))
	for i := range configs.Execs {
		cmn.Debug("%s: exec %d", funcName, i)
		cmn.Debug("%s: exec %d: path: %s", funcName, i, configs.Execs[i].Path)
		cmn.Debug("%s: exec %d: pattern: %s", funcName, i, configs.Execs[i].Pattern)
		// Walk the tree configured in the processor...retrieving the matched files.
		files, err := cmn.WalkMatch(configs.Execs[i].Path, configs.Execs[i].Pattern)
		if err != nil {
			return err
		}

		// Loop through each file matched...
		cmn.Debug("%s: iterating found files: %d", funcName, len(files))
		for j := range files {
			// Process the command in the processor as a template.
			outTemplate, err := template.New("outTemplate").Funcs(sprig.FuncMap()).Parse(configs.Execs[i].Command)
			if err != nil {
				return err
			}

			// Convert the template to output string.
			var templateOut bytes.Buffer
			err = outTemplate.Execute(&templateOut, files[j])
			if err != nil {
				return err
			}
			cmn.Debug("%s: exec %d: file %d: command: %s", funcName, i, j, templateOut.String())

			// Execute the command and grab the output.
			cmn.Debug("%s: executing command", funcName)
			cmd := exec.Command("sh", "-c", templateOut.String())
			cmd.Stdout = os.Stdout
			err = cmd.Run()
			if err != nil {
				return err
			}
		}
	}

	cmn.Debug("%s: end", funcName)
	return nil
}

// gitHead - Process Head mode git log processor.
func gitHead(repo *git.Repository, ref *plumbing.Reference, processor cmn.GitLog) error {
	funcName := "processors.gitHead"
	cmn.Debug("%s: begin", funcName)

	// Grab the HEAD commit.
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return err
	}
	cmn.Debug("%s: head commit: %v", funcName, commit.Hash.String())

	// Grab the HEAD commit stats.
	commitStats, err := commit.Stats()
	if err != nil {
		return err
	}
	cmn.Debug("%s: head commit stats length: %d", funcName, len(commitStats))

	// Process the file in the config as a template to create the file name.
	fileTemplate, err := template.New("fileTemplate").Funcs(sprig.FuncMap()).Parse(processor.File)
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
	cmn.Debug("%s: file: %s", funcName, templateFile.String())

	// Process the output template in the config.
	outTemplate, err := template.New("outTemplate").Funcs(sprig.FuncMap()).Parse(processor.Template)
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
	cmn.Debug("%s: templateOut length: %d", funcName, len(templateOut.String()))

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
	cmn.Debug("%s: created file: %s", funcName, templateFile.String())

	// Write the output to the file.
	bytesWritten, err := outFile.WriteString(templateOut.String())
	if err != nil {
		return err
	}
	cmn.Debug("%s: wrote %d bytes to file", funcName, bytesWritten)

	cmn.Debug("%s: end", funcName)
	return nil
}

// gitEach - Process Each mode git log processor.
func gitEach(repo *git.Repository, ref *plumbing.Reference, processor cmn.GitLog) error {
	funcName := "processors.gitEach"
	cmn.Debug("%s: begin", funcName)

	// Get the commit history in an interator.
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return err
	}
	cmn.Debug("%s: created git history iterator", funcName)

	// Iterate through the commits.
	err = commitIter.ForEach(func(commit *object.Commit) error {
		cmn.Debug("%s: commit %s", funcName, commit.Hash.String()[0:7])
		// Grab the commit stats.
		commitStats, err := commit.Stats()
		if err != nil {
			return err
		}
		defer commitIter.Close()
		cmn.Debug("%s: commit %s: stats length: %d", funcName, commit.Hash.String()[0:7], len(commitStats))

		// Process the file in the config as a template to create the file name.
		fileTemplate, err := template.New("fileTemplate").Funcs(sprig.FuncMap()).Parse(processor.File)
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
		cmn.Debug("%s: commit %s: file: %s", funcName, commit.Hash.String()[0:7], templateFile.String())

		// Process the output template in the config.
		outTemplate, err := template.New("outTemplate").Funcs(sprig.FuncMap()).Parse(processor.Template)
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
		cmn.Debug("%s: commit %s: templateOut length: %d", funcName, commit.Hash.String()[0:7], len(templateOut.String()))

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
		cmn.Debug("%s: commit %s: created file: %s", funcName, commit.Hash.String()[0:7], templateFile.String())

		// Write the output to the file.
		bytesWritten, err := outFile.WriteString(templateOut.String())
		if err != nil {
			return err
		}
		cmn.Debug("%s: commit %s: wrote %d bytes to file", funcName, commit.Hash.String()[0:7], bytesWritten)

		return nil
	})
	if err != nil {
		return err
	}

	cmn.Debug("%s: end", funcName)
	return nil
}

// gitAll - Process All mode git log processor.
func gitAll(repo *git.Repository, ref *plumbing.Reference, processor cmn.GitLog) error {
	funcName := "processors.gitAll"
	cmn.Debug("%s: begin", funcName)

	// Get the commit history in an interator.
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return err
	}
	defer commitIter.Close()
	cmn.Debug("%s: created git history iterator", funcName)

	var allGit cmn.GitAll

	// Grab the HEAD commit.
	allGit.Head.Commit, err = repo.CommitObject(ref.Hash())
	if err != nil {
		return err
	}
	cmn.Debug("%s: head commit: %v", funcName, allGit.Head.Commit.Hash.String())

	allGit.Head.Stats, err = allGit.Head.Commit.Stats()
	if err != nil {
		return err
	}
	cmn.Debug("%s: head commit stats length: %d", funcName, len(allGit.Head.Stats))

	// Iterate through the commits.
	// allGit.Commits = []GitLogEntry{}
	err = commitIter.ForEach(func(commit *object.Commit) error {
		commitStats, err := commit.Stats()
		if err != nil {
			return err
		}
		cmn.Debug("%s: commit %s: stats length: %d", funcName, commit.Hash.String()[0:7], len(commitStats))
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
	fileTemplate, err := template.New("fileTemplate").Funcs(sprig.FuncMap()).Parse(processor.File)
	if err != nil {
		return err
	}
	var templateFile bytes.Buffer
	err = fileTemplate.Execute(&templateFile, allGit)
	if err != nil {
		return err
	}
	cmn.Debug("%s: file: %s", funcName, templateFile.String())

	// Process the output template in the config.
	outTemplate, err := template.New("outTemplate").Funcs(sprig.FuncMap()).Parse(processor.Template)
	if err != nil {
		return err
	}
	var templateOut bytes.Buffer
	err = outTemplate.Execute(&templateOut, allGit)
	if err != nil {
		return err
	}
	cmn.Debug("%s: templateOut length: %d", funcName, len(templateOut.String()))

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
	cmn.Debug("%s: created file: %s", funcName, templateFile.String())

	// Write the output to the file.
	bytesWritten, err := outFile.WriteString(templateOut.String())
	if err != nil {
		return err
	}
	cmn.Debug("%s: wrote %d bytes to file", funcName, bytesWritten)

	cmn.Debug("%s: end", funcName)
	return nil
}

// Gits - Process the configured git log handlers.
func Gits(configs *cmn.Configs) error {
	funcName := "processors.Gits"
	cmn.Debug("%s: begin", funcName)

	// Iterate through the configured git log handlers.
	cmn.Debug("%s: iterating gits: %d", funcName, len(configs.Gits))
	for i := range configs.Gits {
		// Define the path to the git repository.
		var repoPath string
		if configs.Gits[i].Path != "" {
			repoPath = configs.Gits[i].Path
		} else {
			repoPath = "."
		}
		cmn.Debug("%s: repo path: %s", funcName, repoPath)

		// Open the repository.
		repo, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{
			DetectDotGit:          true,
			EnableDotGitCommonDir: true,
		})
		if err != nil {
			return err
		}
		cmn.Debug("%s: repo opened")

		// Get the HEAD commit.
		ref, err := repo.Head()
		if err != nil {
			return err
		}
		cmn.Debug("%s: head commit: %v", funcName, ref.Hash().String())

		// Iterate through the configured processors.
		cmn.Debug("%s: git %d: iterating processors: %d", funcName, i, len(configs.Gits))
		for j := range configs.Gits[i].Processors {
			cmn.Debug("%s: git %d: processor %d", funcName, i, j)
			if configs.Gits[i].Processors[j].Mode != "" {
				// Get the mode.
				switch strings.ToLower(configs.Gits[i].Processors[j].Mode) {
				case "head":
					// Process the HEAD git config.
					cmn.Debug("%s: git %d: processor %d: mode: head", funcName, i, j)
					err := gitHead(repo, ref, configs.Gits[i].Processors[j])
					if err != nil {
						return err
					}
				case "each":
					// Process the Each git config.
					cmn.Debug("%s: git %d: processor %d: mode: each", funcName, i, j)
					err := gitEach(repo, ref, configs.Gits[i].Processors[j])
					if err != nil {
						return err
					}
				case "all":
					// Process the All git config.
					cmn.Debug("%s: git %d: processor %d: mode: all", funcName, i, j)
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

	cmn.Debug("%s: end", funcName)
	return nil
}
