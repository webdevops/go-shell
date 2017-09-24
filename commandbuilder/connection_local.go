package commandbuilder

func (connection *Connection) LocalCommandBuilder(cmd string, args ...string) []interface{} {
	return CommandInterfaceBuilder(cmd, args...)
}
