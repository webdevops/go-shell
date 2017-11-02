# go-shell

[![GitHub release](https://img.shields.io/github/release/webdevops/go-shell.svg)](https://github.com/webdevops/go-shell/releases)
[![license](https://img.shields.io/github/license/webdevops/go-shell.svg)](https://github.com/webdevops/go-shell/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/webdevops/go-shell.svg?branch=master)](https://travis-ci.org/webdevops/go-shell)

Library to write "shelling out" Go code more shell-like,
while remaining idiomatic to Go.

## Features

 * Function-wrapper factories for shell commands
 * Panic on non-zero exits for `set -e` behavior
 * Result of `Run()` is a Stringer for STDOUT, has Error for STDERR
 * Heavily variadic function API `Cmd("rm", "-r", "foo") == Cmd("rm -r", "foo")`
 * Go-native piping `Cmd(...).Pipe(...)` or inline piping `Cmd("... | ...")`
 * Template compatible "last arg" piping `Cmd(..., Cmd(..., Cmd(...)))`
 * Optional trace output mode like `set +x`
 * Similar variadic functions for paths and path templates
 * CommandBuilder for creating command using SSH, Docker or Docker over SSH

## Docker support (commandbuilder)

Running commands inside docker containers can be used by the `CommandBuilder` (see examples below).
The variable `connection.Docker` has to set to the docker container id or the following schema has to be used:

*CONTAINER* is the name of the docker-compose container.

| DSN style configuration                                             | Description                                                                                            |
|:--------------------------------------------------------------------|:-------------------------------------------------------------------------------------------------------|
| ``compose:CONTAINER``                                               | Lookup container id using docker-compose in current directory                                          |
| ``compose:CONTAINER;path=/path/to/project``                         | Lookup container id using docker-compose in `/path/to/project` directory                               |
| ``compose:CONTAINER;path=/path/to/project;file=custom-compose.yml`` | Lookup container id using docker-compose in `/path/to/project` directory and `custom-compose.yml` file |
| ``compose:CONTAINER;project-name=foobar``                           | Lookup container id using docker-compose in current directory with project name `foobar`               |
| ``compose:CONTAINER;host=example.com``                              | Lookup container id using docker-compose in current directory with docker host `example.com`           |
| ``compose:CONTAINER;env[FOOBAR]=BARFOO``                            | Lookup container id using docker-compose in current directory with env var `FOOBAR` set to `BARFOO`    |

| Query style configuration                                             | Description                                                                                            |
|:----------------------------------------------------------------------|:-------------------------------------------------------------------------------------------------------|
| ``compose://CONTAINER``                                               | Lookup container id using docker-compose in current directory                                          |
| ``compose://CONTAINER?path=/path/to/project``                         | Lookup container id using docker-compose in `/path/to/project` directory                               |
| ``compose://CONTAINER?path=/path/to/project&file=custom-compose.yml`` | Lookup container id using docker-compose in `/path/to/project` directory and `custom-compose.yml` file |
| ``compose://CONTAINER?project-name=foobar``                           | Lookup container id using docker-compose in current directory with project name `foobar`               |
| ``compose://CONTAINER?host=example.com``                              | Lookup container id using docker-compose in current directory with docker host `example.com`           |
| ``compose://CONTAINER?env[FOOBAR]=BARFOO``                            | Lookup container id using docker-compose in current directory with env var `FOOBAR` set to `BARFOO`    |


## Examples `shell`

```go
import (
  "fmt"
  "github.com/webdevops/go-shell"
)

var (
  sh = shell.Run
)

shell.Trace = true // like set +x
shell.Shell = []string{"/bin/bash", "-c"} // defaults to /bin/sh

func main() {
  defer shell.ErrExit()
  sh("echo Foobar > /foobar")
  sh("rm /fobar") // typo raises error
  sh("echo Done!") // never run, program exited
}
```

```go
import (
  "fmt"
  "github.com/webdevops/go-shell"
)

func main() {
  cmd := shell.Cmd("echo", "foobar").Pipe("wc", "-c").Pipe("awk", "'{print $1}'")
  
  // -> wc -c | awk '{print $1}'
  fmt.Println(cmd.ToString())
  
  // -> 7 
  fmt.Println(cmd.Run())
}
```

```go
import "github.com/webdevops/go-shell"

var (
  echo = shell.Cmd("echo").OutputFn()
  copy = shell.Cmd("cp").ErrFn()
  rm = shell.Cmd("rm").ErrFn()
)

func main() {
  err := copy("/foo", "/bar")
  // handle err
  err = rm("/bar")
  // handle err
  out, _ := echo("Done!")
}
```

Error recovery
```go
package main

import (
	"os"
	"fmt"
	"github.com/webdevops/go-shell"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("%v", r)

			if obj, ok := r.(*shell.Process); ok {
				message = obj.Debug()
			}
			fmt.Println(message)
			os.Exit(255)
		}
	}()

	shell.Cmd("exit", "2").Run()
}
```

## Examples `commandbuilder`

```go
package main

import (
	"fmt"
	"github.com/webdevops/go-shell"
	"github.com/webdevops/go-shell/commandbuilder"
)

func main() {
	var cmd *shell.Command
	var connection commandbuilder.Connection

	// ------------------------------------------
	// local execution
	connection = commandbuilder.Connection{}
		
	cmd = shell.Cmd(connection.CommandBuilder("date")...)
	cmd.Run()

	// ------------------------------------------
	// SSH access
	connection = commandbuilder.Connection{}
	connection.Ssh.Set("foobar@example.com")
	
	cmd = shell.Cmd(connection.CommandBuilder("date")...)
	fmt.Println("LOCAL: " + cmd.Run().Stdout.String())
	
	
	// ------------------------------------------
	// Docker execution
	connection = commandbuilder.Connection{}
	connection.Docker.Set("32ceb49d2958")
	
	cmd = shell.Cmd(connection.CommandBuilder("date")...)
	fmt.Println("DOCKER: " + cmd.Run().Stdout.String())

	// ------------------------------------------
	// Docker (lookup via docker-compose) execution
	connection = commandbuilder.Connection{}
	connection.Docker.Set("compose:mysql;path=/path/to/project;file=custom-compose-yml")

	cmd = shell.Cmd(connection.CommandBuilder("date")...)
	fmt.Println("DOCKER with COMPOSE: " + cmd.Run().Stdout.String())

	// ------------------------------------------
	// Docker on remote host via SSH execution
	connection = commandbuilder.Connection{}
	connection.Ssh.Set("foobar@example.com")
	connection.Docker.Set("32ceb49d2958")
	
	cmd = shell.Cmd(connection.CommandBuilder("date")...)
	fmt.Println("DOCKER via SSH: " + cmd.Run().Stdout.String())
}
```

## License

MIT
