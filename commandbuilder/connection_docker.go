package commandbuilder

import (
	"strings"
	"fmt"
	"bufio"
	"github.com/webdevops/go-shell"
)

// Checks if connection is using Docker
func (connection *Connection) IsDocker() bool {
	return !connection.Docker.IsEmpty()
}

// Create dockerized command
func (connection *Connection) DockerCommandBuilder(cmd string, args ...string) []interface{} {
	dockerArgs := append(ConnectionDockerArguments, connection.DockerGetContainerId(), cmd)
	dockerArgs = append(dockerArgs, args...)

	if connection.GetType() == "ssh+docker" {
		// docker on remote server
		return connection.SshCommandBuilder("docker", dockerArgs...)
	} else {
		// local docker server
		return connection.LocalCommandBuilder("docker", dockerArgs...)
	}
}

// Detect docker container id with docker-compose support
func (connection *Connection) DockerGetContainerId() string {
	var container string

	cacheKey := fmt.Sprintf("%s:%s", connection.Docker.Scheme, connection.Docker.Hostname)

	container = connection.queryDockerContainerId()
	return container

	if val, ok := connection.containerCache[cacheKey]; ok {
		// use cached container id
		container = val
	} else {
		container = connection.queryDockerContainerId()
	}
	
	// cache value
	connection.containerCache[cacheKey] = container

	return container
}

func (connection *Connection) queryDockerContainerId() string {
	ret := ""

	// copy connection because we need conn without docker usage (endless loop)
	connectionClone := connection.Clone()
	connectionClone.Docker.Clear()
	connectionClone.Type = "auto"

	connectionClone.Environment.AddMap(connection.Docker.Environment)
	
	switch strings.ToLower(connection.Docker.Scheme) {
	case "docker-compose":
		fallthrough
	case "compose":
		// -> trying to get id from docker-compose
		dockerComposeArgs := []string{"--no-ansi"}

		// map path to project-directory
		if val, ok := connection.Docker.Options["path"]; ok {
			connection.Docker.Options["project-directory"] = val
			delete(connection.Docker.Options, "path")
		}

		for varName, varValue := range connection.Docker.Options {
			dockerComposeArgs = append(dockerComposeArgs, "--" + varName, shell.Quote(varValue))
		}

		// docker-compose command with container name
		dockerComposeArgs = append(dockerComposeArgs, "ps", "-q", shell.Quote(connection.Docker.Hostname))

		// query container id from docker-compose
		cmd := shell.Cmd(connectionClone.RawCommandBuilder("docker-compose", dockerComposeArgs...)...).Run()
		ret = strings.TrimSpace(cmd.Stdout.String())

		if ret == "" {
			panic(fmt.Sprintf("Container \"%s\" not found empty", connection.Docker.Hostname))
		}

	case "docker":
		fallthrough
	case "":
		if connection.Docker.Hostname != "" {
			ret = connection.Docker.Hostname
		} else {
			// only docker container id as passed value?
			ret = connection.Docker.Raw
		}

	default:
		panic(fmt.Sprintf("Docker scheme \"%s\" is not supported for: %s", connection.Docker.Scheme, connection.Docker.Hostname))
	}

	return ret
}

// Detect docker container id with docker-compose support
func (connection *Connection) DockerGetEnvironment() map[string]string {
	ret := map[string]string{}

	containerId := connection.DockerGetContainerId()

	// copy connection because we need conn without docker usage (endless loop)
	connectionClone := connection.Clone()
	connectionClone.Docker.Clear()
	connectionClone.Type = "auto"

	cmd := shell.Cmd(connectionClone.CommandBuilder("docker", "inspect", "-f", "{{range .Config.Env}}{{println .}}{{end}}", containerId)...)
	envList := cmd.Run().Stdout.String()

	scanner := bufio.NewScanner(strings.NewReader(envList))
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.SplitN(line, "=", 2)

		if len(split) == 2 {
			varName, varValue := split[0], split[1]

			ret[varName] = varValue
		}
	}

	return ret
}

