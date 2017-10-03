package commandbuilder

import (
	"strings"
	"fmt"
	"github.com/webdevops/go-shell"
)

var containerCache = map[string]string{}

// Create dockerized command
func (connection *Connection) DockerCommandBuilder(cmd string, args ...string) []interface{} {
	containerName := connection.Docker

	dockerArgs := append(ConnectionDockerArguments, connection.DockerGetContainerId(containerName), cmd)
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
func (connection *Connection) DockerGetContainerId(containerName string) string {
	var container string

	cacheKey := fmt.Sprintf("%s:%s", connection.Hostname, containerName)

	if val, ok := containerCache[cacheKey]; ok {
		// use cached container id
		container = val
	} else if strings.HasPrefix(containerName, "compose:") {
		// docker-compose container
		// -> trying to get id from docker-compose

		// copy connection because we need conn without docker usage (endless loop)
		connectionClone := *connection
		connectionClone.Docker = ""
		connectionClone.Type  = "auto"

		// extract docker-compose container name
		composeContainerName := strings.TrimPrefix(containerName, "compose:")

		// query container id from docker-compose
		cmd := shell.Cmd(connectionClone.CommandBuilder("docker-compose", "ps", "-q", composeContainerName)...).Run()
		containerId := strings.TrimSpace(cmd.Stdout.String())

		if containerId == "" {
			panic(fmt.Sprintf("Container \"%s\" not found empty", container))
		}

		container = containerId
	} else {
		// normal docker
		container = containerName
	}

	// cache value
	containerCache[cacheKey] = container

	return container
}
