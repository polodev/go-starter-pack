package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
	"goravel/app/modular"
)

func Commands() []console.Command {
	return append(modular.Collect(Modules(), func(m modular.HasCommands) []console.Command { return m.Commands() }),
		&commands.BackupRun{},
		&commands.BackupList{},
		&commands.QueueWork{},
		&commands.MakeModule{},
	)
}
