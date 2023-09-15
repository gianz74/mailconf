package myterm

import (
	"errors"
	"io"
	"strconv"

	"github.com/gianz74/mailconf/internal/os"
	"golang.org/x/term"
)

var ErrNoTerm = errors.New("Not a terminal.")

var _term Terminal

type Password string

type Terminal interface {
	ReadLine(string) (string, error)
	ReadPass(string) (string, error)
	YesNo(string) bool
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
	ret := &_Term{
		stdin:  stdin,
		stdout: stdout,
	}

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

func (t *_Term) YesNo(prompt string) bool {
	ans, err := t.ReadLine(prompt)
	if err != nil {
		return false
	}

	if len(ans) == 0 {
		ans = "n"
	}
	if ans[0] == 'y' || ans[0] == 'Y' {
		return true
	}
	return false
}

func (t *_Term) ReadPass(prompt string) (string, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	return t.t.ReadPassword(prompt)
}

type Question struct {
	Prompt string
	Var    any
}

func (t *_Term) Ask(questions []Question) {
	for _, q := range questions {
		if _, ok := q.Var.(*string); ok {
			val, _ := t.ReadLine(q.Prompt)
			q.Var = &val
		}
		if _, ok := q.Var.(*Password); ok {
			val, _ := t.ReadPass(q.Prompt)
			q.Var = &val
		}
		if _, ok := q.Var.(*uint16); ok {
			val, _ := t.ReadLine(q.Prompt)
			for {
				port, err := strconv.ParseInt(val, 10, 16)
				if err == nil {
					q.Var = &port
					break
				}
				val, _ = t.ReadLine("please provide a value between 0 and 65535: ")
			}
		}
	}
}
