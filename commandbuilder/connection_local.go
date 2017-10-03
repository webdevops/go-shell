package commandbuilder

// Create local command
func (connection *Connection) LocalCommandBuilder(cmd string, args ...string) []interface{} {
	return CommandInterfaceBuilder(cmd, args...)
}
