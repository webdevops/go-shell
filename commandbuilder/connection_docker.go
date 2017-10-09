package commandbuilder

import (
	"strings"
	"fmt"
	"bufio"
	"regexp"
	"net/url"
	"github.com/webdevops/go-shell"
)

var containerCache = map[string]string{}

type dockerConfiguration struct {
	Scheme string
	Container string
	Options map[string]string
	Environment map[string]string
}

var dockerDsnRegexp = regexp.MustCompile("^(?P<schema>[a-zA-Z0-9]+):(?P<container>[^/;]+)(?P<params>;[^/;]+.*)")
var dockerUrlRegexp = regexp.MustCompile("^[a-zA-Z0-9]+://[^/]+")
var dockerOptionEnvSetting = regexp.MustCompile("^env\\[(?P<item>[a-zA-Z0-9]+)\\]$")

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
	} else if dockerDsnRegexp.MatchString(containerName) {
		container = connection.dockerContainerIdFromDsn(containerName)
	} else if dockerUrlRegexp.MatchString(containerName) {
		container = connection.dockerContainerIdFromUrl(containerName)
	} else {
		// normal docker
		container = containerName
	}

	// cache value
	containerCache[cacheKey] = container

	return container
}

func (connection *Connection) dockerContainerIdFromDsn(containerName string) string {
	docker := dockerConfiguration{}

	match := dockerDsnRegexp.FindStringSubmatch(containerName)
	docker.Scheme = match[1]
	docker.Container = match[2]
	docker.Options = map[string]string{}
	docker.Environment = map[string]string{}

	dockerParams := match[3]
	dockerParams = strings.Trim(dockerParams, ";")
	paramRawList := strings.Split(dockerParams, ";")
	for _, val := range paramRawList {
		split := strings.SplitN(val, "=", 2)
		if len(split) == 2 {
			varName := split[0]
			varValue := split[1]
			if dockerOptionEnvSetting.MatchString(varName) {
				// env value
				match := dockerOptionEnvSetting.FindStringSubmatch(varName)
				envName := match[1]
				docker.Environment[envName] = varValue
			} else {
				// Default option
				docker.Options[varName] = varValue
			}
		}
	}

	return connection.queryDockerContainerId(docker)
}

func (connection *Connection) dockerContainerIdFromUrl(containerName string) string {
	docker := dockerConfiguration{}

	// parse url
	dockerUrl, err := url.Parse(containerName)
	if err != nil {
		panic(err)
	}

	docker.Scheme = dockerUrl.Scheme
	docker.Container = dockerUrl.Hostname()
	docker.Options = map[string]string{}
	docker.Environment = map[string]string{}

	queryVarList := dockerUrl.Query()

	for varName, varValue := range queryVarList {
		if dockerOptionEnvSetting.MatchString(varName) {
			// env value
			match := dockerOptionEnvSetting.FindStringSubmatch(varName)
			envName := match[1]
			docker.Environment[envName] = varValue[0]
		} else {
			// Default option
			docker.Options[varName] = varValue[0]
		}
	}

	return connection.queryDockerContainerId(docker)
}

func (connection *Connection) queryDockerContainerId(docker dockerConfiguration) string {
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

		// set file
		if val, ok := docker.Options["project-name"]; ok {
			dockerComposeArgs = append(dockerComposeArgs, "--project-name", shell.Quote(val))
		}

		// set file
		if val, ok := docker.Options["file"]; ok {
			dockerComposeArgs = append(dockerComposeArgs, "--file", shell.Quote(val))
		}

		// set host
		if val, ok := docker.Options["host"]; ok {
			dockerComposeArgs = append(dockerComposeArgs, "--host", shell.Quote(val))
		}

		// use path/project-dir
		if val, ok := docker.Options["path"]; ok {
			dockerComposeArgs = append(dockerComposeArgs, "--project-directory", shell.Quote(val))
		} else if val, ok := docker.Options["project-directory"]; ok {
			dockerComposeArgs = append(dockerComposeArgs, "--project-directory", shell.Quote(val))
		}

		// docker-compose command with container name
		dockerComposeArgs = append(dockerComposeArgs, "ps", "-q", shell.Quote(docker.Container))

		// query container id from docker-compose
		cmd := shell.Cmd(connectionClone.RawCommandBuilder("docker-compose", dockerComposeArgs...)...).Run()
		ret = strings.TrimSpace(cmd.Stdout.String())

		if ret == "" {
			panic(fmt.Sprintf("Container \"%s\" not found empty", docker.Container))
		}

	default:
		panic(fmt.Sprintf("Docker scheme \"%s\" is not supported for: %s", docker.Scheme, docker.Container))
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

