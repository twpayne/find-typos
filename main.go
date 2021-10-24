package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	format = flag.String("format", "", "format (github-actions or empty)")
	level  = flag.String("level", "warning", "GitHub Action level (notice, warning, or error)")
)

func printTyposInFile(tf *TypoFinder, path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	if !strings.HasPrefix(http.DetectContentType(data), "text/") {
		return 0, nil
	}

	s := bufio.NewScanner(bytes.NewReader(data))
	max := 1024 * 1024
	s.Buffer(make([]byte, max), max)
	lineNumber := 0
	n := 0
	for s.Scan() {
		lineNumber++
		if typos := tf.FindTypos(s.Text()); len(typos) > 0 {
			n += len(typos)
			for _, typo := range typos {
				switch *format {
				case "github-actions":
					if _, err := fmt.Printf("::%s file=%s,line=%d,col=%d::%s: typo of %s\n", *level, path, lineNumber, typo.Index+1, typo.S, tf.word); err != nil {
						return n, err
					}
				default:
					if _, err := fmt.Printf("%s:%d:%d: %s\n", path, lineNumber, typo.Index+1, typo.S); err != nil {
						return n, err
					}
				}
			}
		}
	}
	return n, s.Err()
}

func printTyposInStdin(tf *TypoFinder) (int, error) {
	s := bufio.NewScanner(os.Stdin)
	max := 1024 * 1024
	s.Buffer(make([]byte, max), max)
	lineNumber := 0
	n := 0
	for s.Scan() {
		lineNumber++
		if typos := tf.FindTypos(s.Text()); len(typos) > 0 {
			n += len(typos)
			for _, typo := range typos {
				switch *format {
				case "github-actions":
					if _, err := fmt.Printf("::%s line=%d,col=%d::%s: typo of %s\n", *level, lineNumber, typo.Index+1, typo.S, tf.word); err != nil {
						return n, err
					}
				default:
					if _, err := fmt.Printf("%d:%d: %s\n", lineNumber, typo.Index+1, typo.S); err != nil {
						return n, err
					}
				}
			}
		}
	}
	return n, s.Err()
}

func run() (bool, error) {
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Printf("usage: %s word [path...]\n", filepath.Base(os.Args[0]))
		return false, nil
	}

	tf, err := NewTypoFinder(flag.Arg(0))
	if err != nil {
		return false, err
	}

	if flag.NArg() == 1 {
		n, err := printTyposInStdin(tf)
		if err != nil {
			return false, err
		}
		return n == 0, nil
	}

	total := 0
	fsys := os.DirFS(".")
	for _, arg := range flag.Args()[1:] {
		switch info, err := os.Stat(arg); {
		case err != nil:
			return false, err
		case info.IsDir():
			if err := fs.WalkDir(fsys, arg, func(path string, d fs.DirEntry, err error) error {
				switch {
				case err != nil:
					return err
				case d.IsDir():
					switch {
					case filepath.Base(path) == ".git":
						return fs.SkipDir
					default:
						return nil
					}
				case d.Type() == 0:
					n, err := printTyposInFile(tf, path)
					total += n
					if err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				return false, err
			}
		case info.Mode().IsRegular():
			n, err := printTyposInFile(tf, arg)
			total += n
			if err != nil {
				return false, err
			}
		}
	}

	return total == 0, nil
}

func main() {
	ok, err := run()
	if err != nil {
		fmt.Println(err)
	}
	if !ok || err != nil {
		os.Exit(1)
	}
}
