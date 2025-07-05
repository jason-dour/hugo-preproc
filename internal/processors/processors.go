// Package processors provide the various functions to run processors.
package processors

import (
	"bytes"
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

// panicifError was always a cheap hack; we're going to fix this soon.
func panicIfError(err error) {
	// TODO - Replace all of the calls to this function with proper error handling.
	if err != nil {
		panic(err)
	}
}

// Execs iterates through the exec command processors in the config file.
func Execs(configs *cmn.Configs) {
	// Loop through each processor...
	for i := 0; i < len(configs.Processors); i++ {
		// Walk the tree configured in the processor...retrieving the matched files.
		files, err := cmn.WalkMatch(configs.Processors[i].Path, configs.Processors[i].Pattern)
		panicIfError(err)

		// Loop through each file matched...
		for j := 0; j < len(files); j++ {
			// Process the command in the processor as a template.
			outTemplate, err := template.New("outTemplate").Funcs(cmn.FuncMap).Parse(configs.Processors[i].Command)
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

// gitHead - Process Head mode git log processor.
func gitHead(repo *git.Repository, ref *plumbing.Reference, processor cmn.GitLog) {
	// Grab the HEAD commit.
	commit, err := repo.CommitObject(ref.Hash())
	panicIfError(err)
	commitStats, err := commit.Stats()
	panicIfError(err)

	// Process the file in the config as a template to create the file name.
	fileTemplate, err := template.New("fileTemplate").Funcs(cmn.FuncMap).Parse(processor.File)
	panicIfError(err)
	var templateFile bytes.Buffer
	err = fileTemplate.Execute(&templateFile, cmn.GitLogEntry{
		Commit: commit,
		Stats:  commitStats,
	})
	panicIfError(err)

	// Process the output template in the config.
	outTemplate, err := template.New("outTemplate").Funcs(cmn.FuncMap).Parse(processor.Template)
	panicIfError(err)
	var templateOut bytes.Buffer
	err = outTemplate.Execute(&templateOut, cmn.GitLogEntry{
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

// gitEach - Process Each mode git log processor.
func gitEach(repo *git.Repository, ref *plumbing.Reference, processor cmn.GitLog) {
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
		fileTemplate, err := template.New("fileTemplate").Funcs(cmn.FuncMap).Parse(processor.File)
		panicIfError(err)
		var templateFile bytes.Buffer
		err = fileTemplate.Execute(&templateFile, cmn.GitLogEntry{
			Commit: commit,
			Stats:  commitStats,
		})
		panicIfError(err)

		// Process the output template in the config.
		outTemplate, err := template.New("outTemplate").Funcs(cmn.FuncMap).Parse(processor.Template)
		panicIfError(err)
		var templateOut bytes.Buffer
		err = outTemplate.Execute(&templateOut, cmn.GitLogEntry{
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

// gitAll - Process All mode git log processor.
func gitAll(repo *git.Repository, ref *plumbing.Reference, processor cmn.GitLog) {
	// Get the commit history in an interator.
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	panicIfError(err)
	defer commitIter.Close()

	var allGit cmn.GitAll

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
		allGit.Commits = append(allGit.Commits, cmn.GitLogEntry{
			Commit: commit,
			Stats:  commitStats,
		})
		return nil
	})
	panicIfError(err)

	// Process the file in the config as a template to create the file name.
	fileTemplate, err := template.New("fileTemplate").Funcs(cmn.FuncMap).Parse(processor.File)
	panicIfError(err)
	var templateFile bytes.Buffer
	err = fileTemplate.Execute(&templateFile, allGit)
	panicIfError(err)

	// Process the output template in the config.
	outTemplate, err := template.New("outTemplate").Funcs(cmn.FuncMap).Parse(processor.Template)
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

// Gits - Process the configured git log handlers.
func Gits(configs *cmn.Configs) {
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
		repo, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{
			DetectDotGit:          true,
			EnableDotGitCommonDir: true,
		})
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
					gitHead(repo, ref, configs.Gits[i].Processors[j])
				case "each":
					// Process the Each git config.
					gitEach(repo, ref, configs.Gits[i].Processors[j])
				case "all":
					// Process the All git config.
					gitAll(repo, ref, configs.Gits[i].Processors[j])
				}
			}
		}
	}
}
