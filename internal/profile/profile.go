package profile

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/gianz74/mailconf/internal/base"
	"github.com/gianz74/mailconf/internal/profile/add"
	"github.com/gianz74/mailconf/internal/profile/list"
)

var CmdProfile = &base.Command{
	UsageLine: "profile command",
	Short:     "profile configures one email account",
}

var Commands []*base.Command

func init() {
	CmdProfile.Run = runProfile
	CmdProfile.Commands = []*base.Command{
		list.CmdList,
		add.CmdAdd,
	}
	CmdProfile.Long = tmpl(usageTemplate, CmdProfile.Commands)
}

func runProfile(cmd *base.Command, args []string) error {
	for _, cmd := range cmd.Commands {
		cmd.Flag.Usage = cmd.Usage
		if cmd.Name() == args[0] {
			cmd.Flag.Parse(args[1:])
			args = cmd.Flag.Args()
			return cmd.Run(cmd, args)
		}
	}
	fmt.Println(tmpl(usageTemplate, cmd.Commands))
	return nil
}

func tmpl(text string, data interface{}) string {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace})
	template.Must(t.Parse(text))
	out := &bytes.Buffer{}
	if err := t.Execute(out, data); err != nil {
		panic(err)
	}
	return string(out.Bytes())
}

const usageTemplate = `profile is a subcommand to help create, remove and edit email profiles.

Usage:
	mailconf profile command [arguments]

The commands are:
{{range .}}
	{{.Name | printf "%-11s"}} {{.Short}}{{end}}

Use "mailconf help profile [command]" for more information about a command.`
