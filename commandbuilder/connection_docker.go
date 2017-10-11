package commandbuilder

import (
	"strings"
	"fmt"
	"bufio"
	"github.com/webdevops/go-shell"
)

var containerCache = map[string]string{}

type dockerConfiguration struct {
	Scheme string
	Container string
	Options map[string]string
	Environment map[string]string
}

// Create dockerized command
func (connection *Connection) DockerCommandBuilder(cmd string, args ...string) []interface{} {
	dockerArgs := append(ConnectionDockerArguments, connection.DockerGetContainerId(connection.Docker), cmd)
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
	} else {
		containerInfo, err := ParseArgument(containerName)
		if err != nil {
			panic(err)
		}
		container = connection.queryDockerContainerId(containerInfo)
	}
	
	// cache value
	containerCache[cacheKey] = container

	return container
}

func (connection *Connection) queryDockerContainerId(docker Argument) string {
	ret := ""

	// copy connection because we need conn without docker usage (endless loop)
	connectionClone := connection.Clone()
	connectionClone.Docker = ""
	connectionClone.Type = "auto"

	for envName, envValue := range docker.Environment {
		connectionClone.Environment[envName] = envValue
	}
	
	switch strings.ToLower(docker.Scheme) {
	case "docker-compose":
		fallthrough
	case "compose":
		// -> trying to get id from docker-compose
		dockerComposeArgs := []string{"--no-ansi"}

		// map path to project-directory
		if val, ok := docker.Options["path"]; ok {
			docker.Options["project-directory"] = val
			delete(docker.Options, "path")
		}

		for varName, varValue := range docker.Options {
			dockerComposeArgs = append(dockerComposeArgs, "--" + varName, shell.Quote(varValue))
		}

		// docker-compose command with container name
		dockerComposeArgs = append(dockerComposeArgs, "ps", "-q", shell.Quote(docker.Host))

		// query container id from docker-compose
		cmd := shell.Cmd(connectionClone.RawCommandBuilder("docker-compose", dockerComposeArgs...)...).Run()
		ret = strings.TrimSpace(cmd.Stdout.String())

		if ret == "" {
			panic(fmt.Sprintf("Container \"%s\" not found empty", docker.Host))
		}

	default:
		panic(fmt.Sprintf("Docker scheme \"%s\" is not supported for: %s", docker.Scheme, docker.Host))
	}

	return ret
}

// Detect docker container id with docker-compose support
func (connection *Connection) DockerGetEnvironment(containerId string) map[string]string {
	ret := map[string]string{}

	conn := connection.Clone()
	conn.Docker = ""
	conn.Type  = "auto"

	cmd := shell.Cmd(connection.CommandBuilder("docker", "inspect", "-f", "{{range .Config.Env}}{{println .}}{{end}}", containerId)...)
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

