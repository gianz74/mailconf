package memterm

import (
	"io"
)

type _memterm struct {
	current int
	lines   []string
}

func New() *_memterm {
	return &_memterm{}
}

func (t *_memterm) ReadLine(prompt string) (string, error) {
	if t.current >= 0 && t.current < len(t.lines) {
		line := t.lines[t.current]
		t.current++
		return line, nil
	}
	return "", io.EOF
}

func (t *_memterm) ReadPass(prompt string) (string, error) {
	return t.ReadLine(prompt)
}

func (t *_memterm) SetLines(l []string) {
	t.current = 0
	t.lines = l
}

func (t *_memterm) AddLine(l string) {
	t.current = 0
	t.lines = append(t.lines, l)
}
