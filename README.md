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
processors:
  - path: path/to/top/level
    pattern: "*.md"
    command: echo {{.}}
    # Clearly this example is rather dull; it simply echoes the name of the found file.
```

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

This allows for a reasonably easy way to specify complex commands for processing files prior to the Hugo run.

Other template functions can be added or mapped in as this codebase evolves.
