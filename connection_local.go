package shell

func NewCmd(command string, args ...string) *Command {
	return Cmd(CommandInterfaceBuilder(command, args...)...)
}

func (connection *Connection) LocalCommandBuilder(cmd string, args ...string) []interface{} {
	return CommandInterfaceBuilder(cmd, args...)
}
