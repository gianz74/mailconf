package help

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/gianz74/mailconf/internal/base"
	"github.com/gianz74/mailconf/internal/os"
)

type helpData struct {
	Cmd       *base.Command
	Ancestors []*base.Command
}

func Help(commands []*base.Command, args []string, ancestors []*base.Command) {
	if len(args) == 0 {
		PrintUsage()
		return
	}
	if commands == nil {
		fmt.Fprintf(os.Stderr, "usage: cli help command\n\nToo many arguments given.\n")
		os.Exit(2)
	}

	arg := args[0]

	for _, cmd := range commands {
		if cmd.Name() == arg {
			if len(args) > 1 {
				ancestors = append(ancestors, cmd)
				Help(cmd.Commands, args[1:], ancestors)
				return
			}
			data := helpData{
				Cmd:       cmd,
				Ancestors: ancestors,
			}
			tmpl(helpTemplate, data)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic: %#q. Run 'cli help'.\n", arg)
	os.Exit(2)
}

func PrintUsage() {
	tmpl(usageTemplate, base.Commands)
}

func tmpl(text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace})
	template.Must(t.Parse(text))
	if err := t.Execute(os.Stderr, data); err != nil {
		panic(err)
	}
}

const usageTemplate = `mailconf is tool to configure accounts for mbsync, imapfilter and mu4e
Usage:
	mailconf command [arguments]
The commands are:
{{range .}}
	{{.Name | printf "%-11s"}} {{.Short}}{{end}}
Use "mailconf help [command]" for more information about a command.
`

const helpTemplate = `usage: mailconf {{range .Ancestors}}{{.Name}} {{end}}{{.Cmd.UsageLine}}
{{.Cmd.Long | trim}}
`
