package commandbuilder

import (
	"strings"
	"fmt"
	"github.com/mohae/deepcopy"
)

var (
	// Default SSH options
	ConnectionSshArguments = []string{"-oBatchMode=yes -oPasswordAuthentication=no"}

	// Default Docker options
	ConnectionDockerArguments = []string{"exec", "-i"}
)

type Environment struct {
	Vars map[string]string
}

type Connection struct {
	// Type of command
	Type string

	Ssh Argument
	Docker Argument

	// Working directory for eg. ssh
	Workdir string

	// Environment variables
	Environment Environment

	containerCache map[string]string
}

// Clone connection
func (connection *Connection) Clone() (conn *Connection) {
	conn = deepcopy.Copy(connection).(*Connection)

	if conn.Environment.Vars == nil {
		conn.Environment.Vars = map[string]string{}
	}

	if conn.containerCache == nil {
		conn.containerCache = map[string]string{}
	}

	if conn.Ssh.Options == nil {
		conn.Ssh.Options = map[string]string{}
	}

	if conn.Ssh.Environment == nil {
		conn.Ssh.Environment = map[string]string{}
	}

	if conn.Docker.Options == nil {
		conn.Docker.Options = map[string]string{}
	}

	if conn.Docker.Environment == nil {
		conn.Docker.Environment = map[string]string{}
	}

	return
}

// Check if connection has set any settings
func (connection *Connection) IsEmpty() (status bool) {
	status = false
	if connection.Workdir != "" { return }
	if ! connection.Environment.IsEmpty() { return }
	if ! connection.Ssh.IsEmpty() { return }
	if ! connection.Docker.IsEmpty() { return }

	return true
}

// Set SSH configuration using string (query, dsn, user@host..)
func (connection *Connection) SetSsh(configuration string) error {
	return connection.Ssh.Set(configuration)
}

// Set docker configuration using string (query, dsn, user@host..)
func (connection *Connection) SetDocker(configuration string) error {
	return connection.Docker.Set(configuration)
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
		parts = append(parts, fmt.Sprintf("Docker:%s", connection.Docker.Hostname))
	case "ssh+docker":
		parts = append(parts, fmt.Sprintf("SSH:%s", connection.SshConnectionHostnameString()))
		parts = append(parts, fmt.Sprintf("Docker:%s", connection.Docker.Hostname))
	default:
	}

	return fmt.Sprintf("Connection[%s]", strings.Join(parts[:]," "))
}

// Check if environment is empty
func (env *Environment) IsEmpty() (bool) {
	return len(env.Vars) == 0
}

// Get all environment vars as map
func (env *Environment) GetMap() (map[string]string){
	return env.Vars
}

// Set environment map (absolute)
func (env *Environment) SetMap(vars map[string]string) {
	env.Vars = vars
}

// Set one environment variable
func (env *Environment) Set(name string, value string) {
	env.Vars[name] = value
}

// Add environment map (adds/overwrites)
func (env *Environment) AddMap(vars map[string]string) {
	for name, val := range vars {
		env.Vars[name] = val
	}
}

// Clears environment map
func (env *Environment) Clear() {
	env.Vars = map[string]string{}
}
