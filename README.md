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

First, define a configuration file.  By default, the config filename is `.hugo-preproc.yaml`.  Define a `processors` key whose value is an array of processor objects with `path`, `pattern`, and `command` values.  Such as:

``` yaml
git:
  path: path/to/the/git/repo
  head:
    output: filename-for-output-can-be-a-template
    template: "template to process to put into the output file."
  each:
    output: filename-for-output-can-be-a-template
    template: "template to process to put into the output file."
  all:
    output: filename-for-output-can-be-a-template
    template: "template to process to put into the output file."
processors:
  - path: path/to/top/level
    pattern: "*.md"
    command: echo {{.}}
    # Clearly this example is rather dull; it simply echoes the name of the found file.
```

To enable easy processing of git log data into Hugo pages, there are different git output options that may be used:

* `path` - Defines the path to the git repo (default: ".")
* `head` - For the HEAD commit.
* `each` - For iterating through every commit; executing for each commit.
* `all` - For passing the entire log to the template.

Each `processors` array object is defined as follows:

* `path` - The top-level path that will be walked and scanned for matching filenames.
* `pattern` - The pattern used to match the filenames while walking the `path` contents recursively.
* `command` - The command to run on matching files; this value is processed as a Go template.

When loaded, the configuration informs `hugo-preproc` where to scan, what to match for, and what command to execute for each file matched.  Multiple `processors` entries will be executed serially, in the order in which they are defined.

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
