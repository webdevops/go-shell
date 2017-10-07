package commandbuilder

import (
	"strings"
	"fmt"
	"github.com/webdevops/go-shell"
)

var (
	// Default SSH options
	ConnectionSshArguments = []string{"-oBatchMode=yes -oPasswordAuthentication=no"}

	// Default Docker options
	ConnectionDockerArguments = []string{"exec", "-i"}
)

type Connection struct {
	// Type of command
	// local: for local execution
	// ssh: for execution using ssh
	// docker: for execution using a docker container
	Type string

	// Hostname for ssh execution
	Hostname string

	// Username for ssh execution
	User string

	// Password for ssh execution
	Password string

	// Docker container ID for execution inside docker container
	// compose:name for automatic docker-compose lookup (docker-compose.yml must be in working directory)
	Docker string

	// Working directory for eg. ssh
	WorkDir string

	// Environment variables
	Environment map[string]string
}

// Clone connection
func (connection *Connection) Clone() (Connection) {
	clone := *connection

	if clone.Environment == nil {
		clone.Environment = map[string]string{}
	}

	return clone
}

func (connection *Connection) IsEmpty() bool {
	ret := true

	if connection.Type != "" {
		ret = false
	}

	if connection.Hostname != "" {
		ret = false
	}

	if connection.User != "" {
		ret = false
	}

	if connection.Password != "" {
		ret = false
	}

	if connection.Docker != "" {
		ret = false
	}

	if connection.WorkDir != "" {
		ret = false
	}

	if len(connection.Environment) > 0 {
		ret = false
	}

	return ret
}

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
	if connection.WorkDir != "" || len(connection.Environment) >= 1 {
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

	if connection.WorkDir != "" {
		// prepend cd in front of command to change work dir
		inlineCommand = fmt.Sprintf("cd %s;%s", shell.Quote(connection.WorkDir), inlineCommand)
	}

	if len(connection.Environment) > 0 {
		envList := []string{}
		for envName, envValue := range connection.Environment {
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

		if (connection.Docker != "") && connection.Hostname != "" {
			connection.Type = "ssh+docker"
		} else if connection.Docker != "" {
			connection.Type = "docker"
		} else if connection.Hostname != "" {
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

// Create human readable string representation of command
func (connection *Connection) String() string {
	var parts []string

	connType := connection.GetType()
	parts = append(parts, fmt.Sprintf("Type:%s", connType))

	switch connType {
	case "ssh":
		parts = append(parts, fmt.Sprintf("SSH:%s", connection.SshConnectionHostnameString()))
	case "docker":
		parts = append(parts, fmt.Sprintf("Docker:%s", connection.Docker))
	case "ssh+docker":
		parts = append(parts, fmt.Sprintf("SSH:%s", connection.SshConnectionHostnameString()))
		parts = append(parts, fmt.Sprintf("Docker:%s", connection.Docker))
	default:
	}

	return fmt.Sprintf("Connection[%s]", strings.Join(parts[:]," "))
}
