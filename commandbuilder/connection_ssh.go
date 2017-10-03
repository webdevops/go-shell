package commandbuilder

import (
	"fmt"
	"strings"
	"github.com/webdevops/go-shell"
)

// Create SSH'ed command
func (connection *Connection) SshCommandBuilder(command string, args ...string) []interface{} {
	remoteCmdParts := []string{command}
	for _, val := range args {
		remoteCmdParts = append(remoteCmdParts, val)
	}
	remoteCmd := shell.Quote(strings.Join(remoteCmdParts, " "))

	sshArgs := append(ConnectionSshArguments, connection.SshConnectionHostnameString(), "--", remoteCmd)

	return CommandInterfaceBuilder("ssh", sshArgs...)
}

// Build ssh connection string (eg. user@hostname)
func (connection *Connection) SshConnectionHostnameString() string {
	if connection.User != "" {
		return fmt.Sprintf("%s@%s", connection.User, connection.Hostname)
	} else {
		return connection.Hostname
	}
}
