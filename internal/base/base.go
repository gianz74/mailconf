package base

import (
	"flag"
	"fmt"
	"strings"

	"github.com/gianz74/mailconf/internal/os"
)

type Command struct {
	Run       func(cmd *Command, args []string) error
	UsageLine string
	Short     string
	Long      string
	Flag      flag.FlagSet
	Commands  []*Command
}

var Commands []*Command

func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n", c.UsageLine)
	fmt.Fprintf(os.Stderr, "Run 'mailconf help %s' for details.\n", c.Name())
	os.Exit(2)
}

var Usage func()
