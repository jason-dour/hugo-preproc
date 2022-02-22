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

## Use

First, define a configuration file.  By default, the config filename is `.hugo-preproc.yaml`.  There are two primary keys: `git` and `processors`, such as this example:

``` yaml
git:
  - path: path/to/repo
    processors:
      - mode: head | each | all
        file: path/to/output/file1
        template: |
          Entry {{ . }}
  - path: path/to/repo
    processors:
      - mode: head | each | all
        file: path/to/output/file2
        template: "Entry {{ . }}"
exec:
  - path: path/to/top/directory
    pattern: "*.md"
    command: echo {{ . }}
```

The `git` key  is an array object, with each array element defined as follows:

* `path` - Defines the path to the git repo (default: ".")
* `processors` - Array of git log handlers.
  * `mode` - Values of `head` (only the head commit), `each` (each log entry passed through the processor), or `all` (all entries passed through the processor).
  * `file` - The file to output; processed as a template.
  * `template` - The template through which the git log entry/entries will be processed and then written to `file`.

The `exec` key is an array object, with each array element defined as follows:

* `path` - The top-level path that will be walked and scanned for matching filenames.
* `pattern` - The pattern used to match the filenames while walking the `path` contents recursively.
* `command` - The command to run on matching files; this value is processed as a Go template.

The array entries will be executed serially, in the order in which they are defined.

![Configuration Data Structure](config-data-model.svg)

## Go Templates

We are using Go Templates to process the `command` key in each `processors` object.  This allows for the command to use the matched file name (and derivations of it) as part of the `command`.

Other than standard Go Template functions, we also add:

* `replace` - `strings.Replace`
  * Use: `{{replace <input> <search> <replace> <n>}}`
  * Example:
    * Matched name: `example.drawio`
    * `command`: `draw.io --export --output {{replace . ".md" ".svg" -1}} --format svg {{.}}`
    * Template output used for `exec`: `draw.io --export --output example.svg --format svg example.drawio`
* `split` - `strings.Split`
  * Use: `{{split <input> <separator>}}`
* `trimsuffix` - `strings.TrimSuffix`
  * Use: `{{trimsuffix <input> <trim_string>}}`

This allows for a reasonably easy way to specify complex commands for processing files prior to the Hugo run.

Other template functions can be added or mapped in as this codebase evolves.
