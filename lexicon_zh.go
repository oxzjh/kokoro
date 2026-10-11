package kokoro

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/oxzjh/chinese"
	"github.com/oxzjh/pinyin"
)

var zhWord = regexp.MustCompile(`\p{Han}+`)

type LexiconZH struct {
	lexicon   map[string]string
	extension ILexicon
}

func (l *LexiconZH) SetExtension(extension ILexicon) {
	l.extension = extension
}

func (l *LexiconZH) Phonemic(s string) string {
	s = strings.ReplaceAll(s, "、", ", ")
	s = strings.ReplaceAll(s, "，", ", ")
	s = strings.ReplaceAll(s, "。", ". ")
	s = strings.ReplaceAll(s, "«", " “")
	s = strings.ReplaceAll(s, "»", "” ")
	s = strings.ReplaceAll(s, "《", " “")
	s = strings.ReplaceAll(s, "》", "” ")
	s = strings.ReplaceAll(s, "（", " (")
	s = strings.ReplaceAll(s, "）", ")")
	s = strings.ReplaceAll(s, "！", "! ")
	s = strings.ReplaceAll(s, "：", ": ")
	s = strings.ReplaceAll(s, "；", "; ")
	s = strings.ReplaceAll(s, "？", "? ")
	s = pinyin.Digit2Chinese(s)
	var phonemes []string
	for _, a := range strings.Fields(s) {
		beg := 0
		var sb strings.Builder
		for _, match := range zhWord.FindAllStringIndex(a, -1) {
			end := match[0]
			if end > beg {
				sb.WriteString(l.phonemicNonWord(a[beg:end]))
			}
			beg = match[1]
			sb.WriteString(l.phonemicWord(a[end:beg]))
		}
		if beg < len(a) {
			sb.WriteString(l.phonemicNonWord(a[beg:]))
		}
		phonemes = append(phonemes, sb.String())
	}
	return strings.Join(phonemes, " ")
}

func (l *LexiconZH) phonemicWord(s string) string {
	// var phonemes []string
	var sb strings.Builder
	rs := []rune(s)
	chinese.Traditional2Simplified(rs)
	n := len(rs)
	i := 0
	for i < n {
		j := i + 1
		m := 0
		p := ""
		for {
			f, ok := l.lexicon[string(rs[i:j])]
			if !ok {
				break
			}
			if f != "" {
				m = j
				p = f
			}
			j++
			if j >= n {
				break
			}
		}
		if m == 0 {
			i++
		} else {
			i = m
			// phonemes = append(phonemes, p)
			sb.WriteString(p)
		}
	}
	// return strings.Join(phonemes, " ")
	return sb.String()
}

func (l *LexiconZH) phonemicNonWord(s string) string {
	if l.extension != nil {
		return l.extension.Phonemic(s)
	}
	var sb strings.Builder
	for _, v := range s {
		if _, ok := phonemesDict[v]; ok {
			sb.WriteString(string(v))
		}
	}
	return sb.String()
}

func LoadZH(file string, extension ILexicon) (*LexiconZH, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	br := bufio.NewReader(f)
	lexicon := make(map[string]string, 103927)
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		a := strings.Fields(string(line))
		if len(a) != 2 {
			return nil, fmt.Errorf("parse lexicon file faild, expect 2 fields, got %d fields (%s)", len(a), string(line))
		}
		lexicon[a[0]] = a[1]
		rs := []rune(a[0])
		n := len(rs)
		for i := 1; i < n; i++ {
			f := string(rs[:i])
			if _, ok := lexicon[f]; !ok {
				lexicon[f] = ""
			}
		}
	}
	return &LexiconZH{lexicon, extension}, nil
}
