package myterm

import (
	"errors"
	"io"

	"github.com/gianz74/mailconf/internal/os"
	"golang.org/x/term"
)

var ErrNoTerm = errors.New("Not a terminal.")

var _term Terminal

type Terminal interface {
	ReadLine(string) (string, error)
	ReadPass(string) (string, error)
}

func SetTerm(t Terminal) Terminal {
	ret := _term
	_term = t
	return ret
}

func New() Terminal {
	if _term == nil {
		_term, _ = newTerm(os.Stdin, os.Stdout)
	}
	return _term
}

type _Term struct {
	t      *term.Terminal
	stdin  *os.File
	stdout *os.File
}

func newTerm(stdin *os.File, stdout *os.File) (*_Term, error) {
	if !term.IsTerminal(int(stdin.Fd())) {
		return nil, ErrNoTerm
	}
	screen := struct {
		io.Reader
		io.Writer
	}{stdin, stdout}
	ret := &_Term{}

	ret.t = term.NewTerminal(screen, "")
	return ret, nil
}

func (t *_Term) ReadLine(prompt string) (string, error) {
	oldState, err := term.MakeRaw(int(t.stdin.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(t.stdin.Fd()), oldState)
	t.t.SetPrompt(prompt)
	return t.t.ReadLine()
}

func (t *_Term) ReadPass(prompt string) (string, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	return t.t.ReadPassword(prompt)
}
