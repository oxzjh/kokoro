package kokoro

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
)

var (
	enWord  = regexp.MustCompile(`[a-z']+`)
	enDigit = regexp.MustCompile(`\.?\d+`)
)

type LexiconEN struct {
	isGB           bool
	lexicon        map[string]string
	digitPhonemes  map[rune]string
	digit2Phonemes map[string]string
	symbolPhonemes map[rune]string
}

func (l *LexiconEN) Phonemic(s string) string {
	var phonemes []string
	for _, a := range strings.Fields(s) {
		a = strings.ToLower(a)
		if p, ok := l.lexicon[a]; ok {
			phonemes = append(phonemes, p)
			continue
		}
		beg := 0
		var sb strings.Builder
		for _, match := range enWord.FindAllStringIndex(a, -1) {
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

func (l *LexiconEN) phonemicWord(s string) string {
	if p, ok := l.lexicon[s]; ok {
		return p
	}
	if s[0] == '\'' {
		return "“" + l.phonemicWord(s[1:])
	}
	if s[len(s)-1] == '\'' {
		return l.phonemicWord(s[:len(s)-1]) + "”"
	}
	if p := l.stemS(s); p != "" {
		return p
	}
	if p := l.stemED(s); p != "" {
		return p
	}
	if p := l.stemING(s); p != "" {
		return p
	}
	return l.phonemicDirect(s)
}

func (l *LexiconEN) stemS(s string) string {
	var stem string
	if strings.HasSuffix(s, "ies") {
		stem = l.lexicon[s[:len(s)-3]+"y"]
	} else if strings.HasSuffix(s, "'s") {
		stem = l.lexicon[s[:len(s)-2]]
	} else if strings.HasSuffix(s, "es") {
		var ok bool
		if stem, ok = l.lexicon[s[:len(s)-2]]; !ok {
			stem = l.lexicon[s[:len(s)-1]]
		}
	} else if strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "ss") {
		stem = l.lexicon[s[:len(s)-1]]
	}
	if stem != "" {
		runes := []rune(stem)
		last := string(runes[len(runes)-1])
		if strings.Contains("ptkfθ", last) {
			stem += "s"
		} else if strings.Contains("szʃʒʧʤ", last) {
			if l.isGB {
				stem += "ɪz"
			} else {
				stem += "ᵻz"
			}
		} else {
			stem += "z"
		}
	}
	return stem
}

func (l *LexiconEN) stemED(s string) string {
	var stem string
	if strings.HasSuffix(s, "ied") {
		stem = l.lexicon[s[:len(s)-3]+"y"]
	} else if strings.HasSuffix(s, "ed") && !strings.HasSuffix(s, "eed") {
		var ok bool
		if stem, ok = l.lexicon[s[:len(s)-2]]; !ok {
			stem = l.lexicon[s[:len(s)-1]]
		}
	}
	if stem != "" {
		runes := []rune(stem)
		last := string(runes[len(runes)-1])
		if strings.Contains("pkfθʃsʧ", last) {
			stem += "t"
		} else if last == "d" {
			if l.isGB {
				stem += "ɪd"
			} else {
				stem += "ᵻd"
			}
		} else if last == "t" {
			stem += "d"
		} else if l.isGB || len(runes) < 2 {
			stem += "ɪd"
		} else if strings.Contains("AIOWYiuæɑəɛɪɹʊʌ", string(runes[len(runes)-2])) {
			stem = string(runes[:len(runes)-1]) + "ɾᵻd"
		} else {
			stem += "ᵻd"
		}
	}
	return stem
}

func (l *LexiconEN) stemING(s string) string {
	var stem string
	if strings.HasSuffix(s, "ing") {
		sub := s[:len(s)-3]
		var ok bool
		if stem, ok = l.lexicon[sub]; !ok {
			stem = l.lexicon[sub+"e"]
		}
	}
	if stem != "" {
		runes := []rune(stem)
		last := string(runes[len(runes)-1])
		if l.isGB {
			if strings.Contains("əː", last) {
				return ""
			}
		} else if len(runes) > 1 && last == "t" && strings.Contains("AIOWYiuæɑəɛɪɹʊʌ", string(runes[len(runes)-2])) {
			stem = string(runes[:len(runes)-1]) + "ɾɪŋ"
		} else {
			stem += "ɪŋ"
		}
	}
	return stem
}

// func (l *LexiconEN) phonemicUnknownWord(s string) string {
// 	var sb strings.Builder
// 	for _, c := range s {
// 		if p, ok := l.lexicon[string(c)]; ok {
// 			sb.WriteString(p)
// 		}
// 	}
// 	return sb.String()
// }

func (l *LexiconEN) phonemicNonWord(s string) string {
	beg := 0
	var sb strings.Builder
	for _, match := range enDigit.FindAllStringIndex(s, -1) {
		end := match[0]
		if end > beg {
			sb.WriteString(l.phonemicDirect(s[beg:end]))
		}
		beg = match[1]
		sb.WriteString(l.phonemicDigit(s[end:beg]))
	}
	if beg < len(s) {
		sb.WriteString(l.phonemicDirect(s[beg:]))
	}
	return sb.String()
}

func (l *LexiconEN) phonemicDigit(s string) string {
	if s[0] == '.' {
		return l.lexicon["point"] + l.phonemicDirect(s[1:])
	}
	if s[0] == '0' || len(s) == 1 || len(s) > 4 {
		return l.phonemicDirect(s)
	}
	if len(s) == 2 {
		return l.phonemicDigit2(s)
	}
	if len(s) == 3 {
		return l.digitPhonemes[rune(s[0])] + l.lexicon["hundred"] + l.phonemicDigit2(s[1:])
	}
	if len(s) == 4 {
		if s[1] == '0' && s[2] == '0' && s[3] == '0' {
			return l.digitPhonemes[rune(s[0])] + l.lexicon["thousand"]
		}
		p := l.phonemicDigit2(s[:2])
		if s[2] == '0' {
			if s[3] == '0' {
				return p + l.lexicon["hundred"]
			}
			return p + l.lexicon["o"] + l.digitPhonemes[rune(s[3])]
		}
		return p + l.phonemicDigit2(s[2:])
	}
	return l.phonemicDirect(s)
}

func (l *LexiconEN) phonemicDigit2(s string) string {
	if s[0] == '0' {
		if s[1] == '0' {
			return ""
		}
		return l.symbolPhonemes['&'] + l.digitPhonemes[rune(s[1])]
	}
	if s[0] == '1' {
		return l.digit2Phonemes[s]
	}
	if s[1] == '0' {
		return l.digit2Phonemes[s]
	}
	return l.digit2Phonemes[s[:1]+"0"] + l.digitPhonemes[rune(s[1])]
}

func (l *LexiconEN) phonemicDirect(s string) string {
	var sb strings.Builder
	for i, c := range s {
		switch c {
		case '.':
			if i+1 == len(s) {
				sb.WriteString(".")
			}
		case ',':
			if i+1 == len(s) {
				sb.WriteString(",")
			}
		default:
			if c >= '0' && c <= '9' {
				sb.WriteString(l.digitPhonemes[c])
			} else if p, ok := l.symbolPhonemes[c]; ok {
				sb.WriteString(p)
			} else if _, ok := phonemesDict[c]; ok {
				sb.WriteString(string(c))
			}
		}
	}
	return sb.String()
}

func NewEN(lexicon map[string]string, isGB bool) *LexiconEN {
	return &LexiconEN{isGB, lexicon, map[rune]string{
		'0': lexicon["zero"],
		'1': lexicon["one"],
		'2': lexicon["two"],
		'3': lexicon["three"],
		'4': lexicon["four"],
		'5': lexicon["five"],
		'6': lexicon["six"],
		'7': lexicon["seven"],
		'8': lexicon["eight"],
		'9': lexicon["nine"],
	}, map[string]string{
		"10": lexicon["ten"],
		"11": lexicon["eleven"],
		"12": lexicon["twelve"],
		"13": lexicon["thirteen"],
		"14": lexicon["fourteen"],
		"15": lexicon["fifteen"],
		"16": lexicon["sixteen"],
		"17": lexicon["seventeen"],
		"18": lexicon["eighteen"],
		"19": lexicon["nineteen"],
		"20": lexicon["twenty"],
		"30": lexicon["thirty"],
		"40": lexicon["forty"],
		"50": lexicon["fifty"],
		"60": lexicon["sixty"],
		"70": lexicon["seventy"],
		"80": lexicon["eighty"],
		"90": lexicon["ninety"],
	}, map[rune]string{
		'%': lexicon["percent"],
		'+': lexicon["plus"],
		'&': lexicon["and"],
		'@': lexicon["at"],
	}}
}

func LoadEN(file string, isGB bool) (*LexiconEN, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m := make(map[string]string, 190481)
	if err = json.NewDecoder(f).Decode(&m); err != nil {
		return nil, err
	}
	return NewEN(m, isGB), nil
}
