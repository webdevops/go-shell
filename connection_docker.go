package shell

import (
	"strings"
	"fmt"
)

var containerCache = map[string]string{}

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

func (connection *Connection) DockerGetContainerId() string {
	var container string

	cacheKey := fmt.Sprintf("%s:%s", connection.Hostname, connection.Docker)

	if val, ok := containerCache[cacheKey]; ok {
		// use cached container id
		container = val
	} else if strings.HasPrefix(connection.Docker, "compose:") {
		// docker-compose container
		// -> trying to get id from docker-compose

		// copy connection because we need conn without docker usage (endless loop)
		connectionClone := *connection
		connectionClone.Docker = ""
		connectionClone.Type  = "auto"

		// extract docker-compose container name
		containerName := strings.TrimPrefix(connection.Docker, "compose:")

		// query container id from docker-compose
		cmd := Cmd(connectionClone.CommandBuilder("docker-compose", "ps", "-q", containerName)...).Run()
		containerId := strings.TrimSpace(cmd.Stdout.String())

		if containerId == "" {
			panic(fmt.Sprintf("Container \"%s\" not found empty", container))
		}

		container = containerId
	} else {
		// normal docker
		container = connection.Docker
	}

	// cache value
	containerCache[cacheKey] = container

	return container
}
