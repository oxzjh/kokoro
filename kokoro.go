package kokoro

import (
	"errors"
	"io"
	"os"
	"unsafe"
)

const (
	SampleRate = 24000

	minTokens = 10
	maxTokens = 510
)

type ILexicon interface {
	Phonemic(string) string
}

type IInterpreter interface {
	Infer([]int64, []float32, float32) ([]float32, error)
	Close()
}

type Kokoro struct {
	styles      map[string][][]float32
	interpreter IInterpreter
}

func (k *Kokoro) Extract(phonemes []rune, style string, speed float32) ([][]float32, error) {
	var waveforms [][]float32
	err := readTokens(phonemes, func(tokens []int64) error {
		w, err := k.infer(tokens, style, speed)
		if err != nil {
			return err
		}
		waveforms = append(waveforms, w)
		return nil
	})
	return waveforms, err
}

func (k *Kokoro) ExtractOneByOne(phonemes []rune, style string, speed float32, callback func([]float32)) error {
	return readTokens(phonemes, func(tokens []int64) error {
		w, err := k.infer(tokens, style, speed)
		if err != nil {
			return err
		}
		callback(w)
		return nil
	})
}

func (k *Kokoro) infer(tokens []int64, style string, speed float32) ([]float32, error) {
	styles, ok := k.styles[style]
	if !ok {
		return nil, errors.New("voice style not exist: " + style)
	}
	styleBlob := styles[len(tokens)-1]
	return k.interpreter.Infer(tokens, styleBlob, speed)
}

func (k *Kokoro) Close() {
	k.interpreter.Close()
}

func New(voices map[string]string, interpreter IInterpreter) (*Kokoro, error) {
	styles := make(map[string][][]float32, len(voices))
	for style, file := range voices {
		voice, err := readVoice(file)
		if err != nil {
			return nil, err
		}
		styles[style] = voice
	}
	return &Kokoro{styles, interpreter}, nil
}

func readVoice(file string) ([][]float32, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	voices := make([][]float32, 0, maxTokens)
	for {
		b := make([]byte, 1024)
		if _, err := io.ReadFull(f, b); err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		voices = append(voices, unsafe.Slice((*float32)(unsafe.Pointer(&b[0])), 256))
	}
	return voices, nil
}

func readTokens(phonemes []rune, callback func([]int64) error) (err error) {
	tokens := make([]int64, len(phonemes))
	for i, v := range phonemes {
		tokens[i] = phonemesDict[v]
	}
	for len(tokens) > maxTokens {
		idx := maxTokens
		for i := maxTokens - 1; i >= minTokens; i-- {
			v := tokens[i]
			if v == 4 || v == 5 || v == 6 {
				idx = i + 1
				break
			}
		}
		if err = callback(tokens[:idx]); err != nil {
			return
		}
		tokens = tokens[idx:]
	}
	if len(tokens) > 0 {
		err = callback(tokens)
	}
	return
}
