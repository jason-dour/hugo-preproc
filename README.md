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

The `git` key defines how to handle processing of git log data into Hugo files.  The `git` parameters are as follows:

* `path` - Defines the path to the git repo (default: ".")
* `head` - For processing the HEAD commit. Only the HEAD commit will be read and passed through the processor.
  * `file` - The file to output; processed as a template.
  * `template` - The template through which the git log entry/entries will be processed and then written to `file`.
* `each` - For iterating through every commit; executing the filename and output templates for each commit.
  * Same as `head`.
* `all` - For passing the entire git log to the template in a large structure.
  * Same as `head`.

The `processors` key is an array object, with each array element defined as follows:

* `path` - The top-level path that will be walked and scanned for matching filenames.
* `pattern` - The pattern used to match the filenames while walking the `path` contents recursively.
* `command` - The command to run on matching files; this value is processed as a Go template.

The array entries will be executed serially, in the order in which they are defined.

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
