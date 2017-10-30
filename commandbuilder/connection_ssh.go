package commandbuilder

import (
	"fmt"
	"strings"
	"github.com/webdevops/go-shell"
)

// Checks if connection is using SSH
func (connection *Connection) IsSsh() bool {
	return !connection.Ssh.IsEmpty()
}

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
	if connection.Ssh.Username != "" {
		return fmt.Sprintf("%s@%s", connection.Ssh.Username, connection.Ssh.Hostname)
	} else {
		return connection.Ssh.Hostname
	}
}
