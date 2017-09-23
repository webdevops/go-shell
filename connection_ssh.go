package shell

import (
	"fmt"
	"strings"
)

func (connection *Connection) SshCommandBuilder(command string, args ...string) []interface{} {
	remoteCmdParts := []string{command}
	for _, val := range args {
		remoteCmdParts = append(remoteCmdParts, val)
	}
	remoteCmd := Quote(strings.Join(remoteCmdParts, " "))

	sshArgs := append(ConnectionSshArguments, connection.SshConnectionHostnameString(), "--", remoteCmd)

	return CommandInterfaceBuilder("ssh", sshArgs...)
}

func (connection *Connection) SshConnectionHostnameString() string {
	if connection.User != "" {
		return fmt.Sprintf("%s@%s", connection.User, connection.Hostname)
	} else {
		return connection.Hostname
	}
}
