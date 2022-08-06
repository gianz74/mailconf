package list

import (
	"fmt"

	"github.com/gianz74/mailconf/internal/base"
)

var CmdList = &base.Command{
	UsageLine: "list",
	Short:     "list profiles.",
	Long: `
List all profiles defined so far.`,
}

func init() {
	CmdList.Run = runList
}

func runList(cmd *base.Command, arg []string) error {
	fmt.Printf("list command speaking!\n")
	return nil
}
