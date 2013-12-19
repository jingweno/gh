package commands

import (
	"github.com/jingweno/gh/utils"
	"os"
)

var cmdUpdate = &Command{
	Run:   update,
	Usage: "update",
	Short: "Update gh",
	Long: `Update gh to the latest version.

Examples:
  git update
`,
}

func update(cmd *Command, args *Args) {
	updater := NewUpdater()
	err := updater.Update()
	utils.Check(err)
	os.Exit(0)
}
