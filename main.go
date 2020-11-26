package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func appendAddRegexpStrs(regexpStrs []string, runes []rune) []string {
	for i := 1; i < len(runes); i++ {
		regexpStr := string(runes[:i]) + `[A-Za-z]` + string(runes[i:])
		regexpStrs = append(regexpStrs, regexpStr)
	}
	return regexpStrs
}

func appendDeleteRegexpStrs(regexpStrs []string, runes []rune) []string {
	for i := range runes {
		regexpStr := string(runes[:i]) + string(runes[i+1:])
		regexpStrs = append(regexpStrs, regexpStr)
	}
	return regexpStrs
}

func appendReplaceRegexpStrs(regexpStrs []string, runes []rune) []string {
	for i, r := range runes {
		regexpStr := string(runes[:i]) + `\B[^` + string(r) + `]\B` + string(runes[i+1:])
		regexpStrs = append(regexpStrs, regexpStr)
	}
	return regexpStrs
}

func appendTransposeRegexpStrs(regexpStrs []string, runes []rune) []string {
	for i := 1; i < len(runes); i++ {
		regexpStr := string(runes[:i-1]) + string(runes[i]) + string(runes[i-1]) + string(runes[i+1:])
		regexpStrs = append(regexpStrs, regexpStr)
	}
	return regexpStrs
}

func run() error {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Printf("usage: %s name [path...]\n", filepath.Base(os.Args[0]))
		return nil
	}
	name := []rune(flag.Arg(0))
	var paths []string
	if flag.NArg() < 2 {
		paths = []string{"."}
	} else {
		paths = flag.Args()[1:]
	}

	var regexpStrs []string
	regexpStrs = appendAddRegexpStrs(regexpStrs, name)
	regexpStrs = appendDeleteRegexpStrs(regexpStrs, name)
	regexpStrs = appendReplaceRegexpStrs(regexpStrs, name)
	regexpStrs = appendTransposeRegexpStrs(regexpStrs, name)

	var sb strings.Builder
	sb.WriteString(`(?i)\b(?:`)
	for i, regexpStr := range regexpStrs {
		if i > 0 {
			sb.WriteByte('|')
		}
		sb.WriteString(regexpStr)
	}
	sb.WriteString(`)\b`)

	re := regexp.MustCompile(sb.String())

	for _, root := range paths {
		if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filepath.Base(path) == ".git" && info.IsDir() {
				return filepath.SkipDir
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			filename := filepath.Join(root, path)
			data, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}

			if !strings.HasPrefix(http.DetectContentType(data), "text/") {
				return nil
			}

			s := bufio.NewScanner(bytes.NewReader(data))
			line := 0
			for s.Scan() {
				line++
				m := re.FindAllString(s.Text(), -1)
				if m == nil {
					continue
				}
				fmt.Printf("%s:%d: %s\n", filename, line, strings.Join(m, ","))
			}
			if err := s.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s\n", filename, err)
			}

			return nil
		}); err != nil {
			return err
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
