package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var wordRx = regexp.MustCompile(`(?i)\A[a-z]{3,}\z`)

type TypoFinder struct {
	word       string
	wordTypoRx *regexp.Regexp
}

func NewTypoFinder(word string) (*TypoFinder, error) {
	if !wordRx.MatchString(word) {
		return nil, fmt.Errorf("invalid word: %q", word)
	}

	runes := []rune(word)
	var regexpStrs []string
	regexpStrs = appendAddRegexpStrs(regexpStrs, runes)
	regexpStrs = appendDeleteRegexpStrs(regexpStrs, runes)
	regexpStrs = appendReplaceRegexpStrs(regexpStrs, runes)
	regexpStrs = appendTransposeRegexpStrs(regexpStrs, runes)

	var sb strings.Builder
	sb.WriteString(`(?i)\b(?:`)
	for i, regexpStr := range regexpStrs {
		if i > 0 {
			sb.WriteByte('|')
		}
		sb.WriteString(regexpStr)
	}
	sb.WriteString(`)\b`)

	return &TypoFinder{
		word:       word,
		wordTypoRx: regexp.MustCompile(sb.String()),
	}, nil
}

func (tf *TypoFinder) FindTypos(s string) []string {
	m := tf.wordTypoRx.FindAllString(s, -1)
	if len(m) == 0 {
		return nil
	}

	uniqueGenuineTypos := make(map[string]struct{})
	for _, typo := range m {
		typo = strings.ToLower(typo)
		if typo == tf.word { // Remove false positives.
			continue
		}
		uniqueGenuineTypos[typo] = struct{}{}
	}
	if len(uniqueGenuineTypos) == 0 {
		return nil
	}

	sortedTypos := make([]string, 0, len(uniqueGenuineTypos))
	for typo := range uniqueGenuineTypos {
		sortedTypos = append(sortedTypos, typo)
	}
	sort.Strings(sortedTypos)
	return sortedTypos
}

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
