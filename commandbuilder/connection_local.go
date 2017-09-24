package commandbuilder

import (
	"github.com/webdevops/go-shell"
)

func NewCmd(command string, args ...string) *shell.Command {
	return shell.Cmd(CommandInterfaceBuilder(command, args...)...)
}

func (connection *Connection) LocalCommandBuilder(cmd string, args ...string) []interface{} {
	return CommandInterfaceBuilder(cmd, args...)
}
