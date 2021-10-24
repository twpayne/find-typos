# findtypos

findtypos finds typos of a single word. It is primarily useful for developers
who want to find typos of their project's name in their codebase.

## Install

    go install github.com/twpayne/findtypos@latest

## Usage

    findtypos word [path...]

This will print the filename and line number of minor misspellings of *word* in
each *path* specified.

Mispellings found include the replacement of a single character, the addition of
a single character or the removal of a character, and the transposition of two
adjacent characters.

If a *path* is a directory, it is recursed into. Directories called `.git` are
ignored, as are non-text files (i.e. files for whose contents
[`net/http.DetectContentType()`](https://pkg.go.dev/net/http#DetectContentType)
returns a content type that does not begin with `text/`)

If no *path*s are specified, then findtypos reads from the standard input.

## Example

Given the input file `example.txt`:

```
This is an example input file for findtypos. It contains a few typos.

fyndtypos finds a number of minor mis-spellings of a single word, including, for
example, the subsitution of a single letter, or the insertion of a single
letter, like finddtypos.
```

Running findtypos prints:

```console
$ findtypos findtypos example.txt
example.txt:3:1: fyndtypos
example.txt:5:14: finddtypos
```

In general, you would run findtypos in the root of your project as a CI step,
for example:

    findtypos myprojectsweirdname .

## License

MIT