package commandbuilder

import (
	"strings"
	"fmt"
	"github.com/webdevops/go-shell"
)

// Build command for shell.Cmd usage
// will automatically check if SSH'ed or docker exec will be used
func (connection *Connection) CommandBuilder(command string, args ...string) []interface{} {
	args = shell.QuoteValues(args...)
	return connection.RawCommandBuilder(command, args...)
}

// Build raw command (not automatically quoted) for shell.Cmd usage
// will automatically check if SSH'ed or docker exec will be used
func (connection *Connection) RawCommandBuilder(command string, args ...string) []interface{} {
	var ret []interface{}

	// if workdir is set
	// use shell'ed command builder
	if connection.Workdir != "" || !connection.Environment.IsEmpty() {
		shellArgs := []string{command}
		shellArgs = append(shellArgs, args...)
		return connection.RawShellCommandBuilder(shellArgs...)
	}

	switch connection.GetType() {
	case "local":
		ret = connection.LocalCommandBuilder(command, args...)
	case "ssh":
		ret = connection.SshCommandBuilder(command, args...)
	case "ssh+docker":
		fallthrough
	case "docker":
		ret = connection.DockerCommandBuilder(command, args...)
	default:
		panic(connection)
	}

	return ret
}

// Run command using an shell (eg. for running pipes or multiple commands)
// will automatically check if SSH'ed or docker exec will be used
func (connection *Connection) ShellCommandBuilder(args ...string) []interface{} {
	args = shell.QuoteValues(args...)
	return connection.RawShellCommandBuilder(args...)
}


// Run raw command (not automatically quoted) using an shell (eg. for running pipes or multiple commands)
// will automatically check if SSH'ed or docker exec will be used
func (connection *Connection) RawShellCommandBuilder(args ...string) []interface{} {
	var ret []interface{}

	inlineArgs := []string{}

	for _, val := range args {
		inlineArgs = append(inlineArgs, val)
	}

	inlineCommand := strings.Join(inlineArgs, " ")

	if connection.Workdir != "" {
		// prepend cd in front of command to change work dir
		inlineCommand = fmt.Sprintf("cd %s;%s", shell.Quote(connection.Workdir), inlineCommand)
	}

	if !connection.Environment.IsEmpty() {
		envList := []string{}
		for envName, envValue := range connection.Environment.GetMap() {
			envList = append(envList, fmt.Sprintf("%s=%s", envName, shell.Quote(envValue)))
		}
		inlineCommand = fmt.Sprintf("export %s;%s", strings.Join(envList, " "), inlineCommand)
	}

	// pipefail emulation
	//inlineCommand += `;echo "${PIPESTATUS[@]}"; for x in "${PIPESTATUS[@]}";do if [ "$x" -ne 0 ];then exit "$x";fi;done;`

	inlineCommand = shell.Quote(inlineCommand)


	switch connection.GetType() {
	case "local":
		ret = connection.LocalCommandBuilder(shell.Shell[0], append(shell.Shell[1:], inlineCommand)...)
	case "ssh":
		ret = connection.SshCommandBuilder(shell.Shell[0], append(shell.Shell[1:], inlineCommand)...)
	case "ssh+docker":
		fallthrough
	case "docker":
		ret = connection.DockerCommandBuilder(shell.Shell[0], append(shell.Shell[1:], inlineCommand)...)
	default:
		panic(connection)
	}

	return ret
}

// Return type of connection, will guess type based on settings if type is empty
func (connection *Connection) GetType() string {
	var connType string

	// autodetection
	if (connection.Type == "") || (connection.Type == "auto") {
		connection.Type = "local"

		if (!connection.Docker.IsEmpty()) && (!connection.Ssh.IsEmpty()) {
			connection.Type = "ssh+docker"
		} else if ! connection.Docker.IsEmpty() {
			connection.Type = "docker"
		} else if ! connection.Ssh.IsEmpty() {
			connection.Type = "ssh"
		}
	}

	switch connection.Type {
	case "local":
		connType = "local"
	case "ssh":
		connType = "ssh"
	case "docker":
		connType = "docker"
	case "ssh+docker":
		connType = "ssh+docker"
	default:
		panic(fmt.Sprint("Unknown connection type \"%s\"", connType))
	}

	return connType
}
