# hugo-preproc

Helper for Hugo to provide for pre-processing of files.

## Purpose

Provide for a flexible pre-processor for Hugo, since we cannot as a community appear to be able to get certain filetypes supported for external handlers/processors in the core Hugo code.

Intended to assist with any sort of pre-processing desired for publishing files, such as:

* Diagrams converted to SVG.
  * Mermaid
  * Draw.io
  * Graphviz
  * etc.
* Generate Markdown or data files for Hugo from Git history.

## Use

A configuration file is used to define processing.  By default, the config filename is `.hugo-preproc.yaml` (or `.toml`, or `.json`).

You can specify a config file on command line with the `-c`/`--config` option.

Execute the command and processing occurs based on the configuration.

## Configuration Syntax

The file has two primary keys: `git` and `processors`, such as this example:

``` yaml
git:
  - path: path/to/repo
    processors:
      - mode: head | each | all
        file: path/to/output/{{ .Commit.Hash }}
        template: Entry {{ .<field> }}
exec:
  - path: path/to/top/directory
    pattern: "*.md"
    command: echo {{ . }}
```

The `git` key  is an array object, with each array element defined as follows:

* `path` - Defines the path to the git repo (default: ".")
* `processors` - Array of git log handlers.
  * `mode` - Values of `head` (only the head commit), `each` (each log entry passed through the processor, consecutively), or `all` (all entries passed through the processor).
  * `file` - The file to output; processed as a template.
  * `template` - The template through which the git log entry/entries will be processed and then written to `file`.

The `exec` key is an array object, with each array element defined as follows:

* `path` - The top-level path that will be walked and scanned for matching filenames.
* `pattern` - The pattern used to match the filenames while walking the `path` contents recursively.
* `command` - The command to run on matching files; this value is processed as a Go template.

The array entries will be executed serially, in the order in which they are defined.

![Configuration Data Structure](config-data-model.drawio.svg)

## Go Templates

We are using Go Templates to process the `file` and `template` keys in each `git` handler, as well as the `command` key in each `processors` object.

We've now mapped the full library of [Masterminds/sprig](https://github.com/Masterminds/sprig) template functions.
As of `2.x` releases, all prior custom functions are deprecated.

## Go Template Input

We provide the following input for the configured handlers.

* `git` handlers
  * `head` and `each`

    ``` go
    . {
      Commit {
        Hash         string   // Hash of the commit object.
        Author       {        // Author is the original author of the commit.
          Name  string    // Name of the Author.
          Email string    // Email address of the Author.
          When  time.Time // Date/time of the commit. 
        }
        Committer    {        // Committer is the one performing the commit,
                              // might be different from Author.
          Name  string    // Name of the Committer.
          Email string    // Email address of the Committer.
          When  time.Time // Date/time of the commit. 
        }
        Message      string   // Message is the commit message, contains arbitrary text.
        TreeHash     string   // TreeHash is the hash of the root tree of the commit.
        ParentHashes []string // ParentHashes are the hashes of the parent commits of the commit.
        PGPSignature string   // PGPSignature is the PGP signature of the commit.
      }
      Stats []string // Array of strings for files changed and their stats.
    }
    ```

  * `all`

    ``` go
    . {
      // Array of Commits
      Commits []{
        Commit {
          Hash         string   // Hash of the commit object.
          Author       string   // Author is the original author of the commit.
            Name  string    // Name of the Author.
            Email string    // Email address of the Author.
            When  time.Time // Date/time of the commit. 
          }
          Committer    {        // Committer is the one performing the commit,
                                // might be different from Author.
            Name  string    // Name of the Committer.
            Email string    // Email address of the Committer.
            When  time.Time // Date/time of the commit. 
          }
          Message      string   // Message is the commit message, contains arbitrary text.
          TreeHash     string   // TreeHash is the hash of the root tree of the commit.
          ParentHashes []string // ParentHashes are the hashes of the parent commits of the commit.
          PGPSignature string   // PGPSignature is the PGP signature of the commit.
        }
        Stats []string // Array of strings for files changed and their stats.
      }
      // Head Commit
      Head {
        Commit {
          Hash         string   // Hash of the commit object.
          Author       string   // Author is the original author of the commit.
            Name  string    // Name of the Author.
            Email string    // Email address of the Author.
            When  time.Time // Date/time of the commit. 
          }
          Committer    {        // Committer is the one performing the commit,
                                // might be different from Author.
            Name  string    // Name of the Committer.
            Email string    // Email address of the Committer.
            When  time.Time // Date/time of the commit. 
          }
          Message      string   // Message is the commit message, contains arbitrary text.
          TreeHash     string   // TreeHash is the hash of the root tree of the commit.
          ParentHashes []string // ParentHashes are the hashes of the parent commits of the commit.
          PGPSignature string   // PGPSignature is the PGP signature of the commit.
        }
        Stats []string // Array of strings for files changed and their stats.
      }
    }
    ```

* `exec` handlers

  ``` go
  . string // String representing the matched filename,
           // including its full path (handler top search path + sub-path to file).
  ```
