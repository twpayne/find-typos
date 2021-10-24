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
)

func printTyposInFile(tf *TypoFinder, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(http.DetectContentType(data), "text/") {
		return nil
	}

	s := bufio.NewScanner(bytes.NewReader(data))
	max := 1024 * 1024
	s.Buffer(make([]byte, max), max)
	lineNumber := 0
	for s.Scan() {
		lineNumber++
		if typos := tf.FindTypos(s.Text()); len(typos) > 0 {
			for _, typo := range typos {
				switch *format {
				case "github-actions":
					if _, err := fmt.Printf("::warning file=%s,line=%d,col=%d::%s: typo of %s\n", path, lineNumber, typo.Index+1, typo.S, tf.word); err != nil {
						return err
					}
				default:
					if _, err := fmt.Printf("%s:%d:%d: %s\n", path, lineNumber, typo.Index+1, typo.S); err != nil {
						return err
					}
				}
			}
		}
	}
	return s.Err()
}

func printTyposInStdin(tf *TypoFinder) error {
	s := bufio.NewScanner(os.Stdin)
	max := 1024 * 1024
	s.Buffer(make([]byte, max), max)
	lineNumber := 0
	for s.Scan() {
		lineNumber++
		if typos := tf.FindTypos(s.Text()); len(typos) > 0 {
			for _, typo := range typos {
				switch *format {
				case "github-actions":
					if _, err := fmt.Printf("::warning line=%d,col=%d::%s: typo of %s\n", lineNumber, typo.Index+1, typo.S, tf.word); err != nil {
						return err
					}
				default:
					if _, err := fmt.Printf("%d:%d: %s\n", lineNumber, typo.Index+1, typo.S); err != nil {
						return err
					}
				}
			}
		}
	}
	return s.Err()
}

func run() error {
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Printf("usage: %s word [path...]\n", filepath.Base(os.Args[0]))
		return nil
	}

	tf, err := NewTypoFinder(flag.Arg(0))
	if err != nil {
		return err
	}

	if flag.NArg() == 1 {
		return printTyposInStdin(tf)
	}

	fsys := os.DirFS(".")
	for _, arg := range flag.Args()[1:] {
		switch info, err := os.Stat(arg); {
		case err != nil:
			return err
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
					if err := printTyposInFile(tf, path); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				return err
			}
		case info.Mode().IsRegular():
			if err := printTyposInFile(tf, arg); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
