package commandbuilder

import (
	"testing"
	"github.com/webdevops/go-shell"
)

func TestConnectionLocal(t *testing.T) {
	var cmd *shell.Command
	conn := Connection{}

	cmd = shell.Cmd(conn.CommandBuilder("echo", "foobar")...)
	if val := cmd.ToString(); val != "echo 'foobar'" {
		t.Fatal("command builder not expected command:", val)
	}

	cmd = shell.Cmd(conn.RawCommandBuilder("echo", "foobar")...)
	if val := cmd.ToString(); val != "echo foobar" {
		t.Fatal("command builder not expected command:", val)
	}
}

func TestConnectionSsh(t *testing.T) {
	var cmd *shell.Command
	conn := Connection{
		Hostname: "example.com",
		User: "barfoo",
	}

	cmd = shell.Cmd(conn.CommandBuilder("echo", "foobar")...)
	if val := cmd.ToString(); val != "ssh -oBatchMode=yes -oPasswordAuthentication=no barfoo@example.com -- 'echo '\\''foobar'\\'''" {
		t.Fatal("command builder not expected command:", val)
	}

	cmd = shell.Cmd(conn.RawCommandBuilder("echo", "foobar")...)
	if val := cmd.ToString(); val != "ssh -oBatchMode=yes -oPasswordAuthentication=no barfoo@example.com -- 'echo foobar'" {
		t.Fatal("command builder not expected command:", val)
	}
}

func TestConnectionDocker(t *testing.T) {
	var cmd *shell.Command
	conn := Connection{
		Docker: "containerid",
	}

	cmd = shell.Cmd(conn.CommandBuilder("echo", "foobar")...)
	if val := cmd.ToString(); val != "docker exec -i containerid echo 'foobar'" {
		t.Fatal("command builder not expected command:", val)
	}

	cmd = shell.Cmd(conn.RawCommandBuilder("echo", "foobar")...)
	if val := cmd.ToString(); val != "docker exec -i containerid echo foobar" {
		t.Fatal("command builder not expected command:", val)
	}
}


func TestConnectionSshDocker(t *testing.T) {
	var cmd *shell.Command
	conn := Connection{
		Hostname: "example.com",
		User: "barfoo",
		Docker: "containerid",
	}

	cmd = shell.Cmd(conn.CommandBuilder("echo", "foobar")...)
	if val := cmd.ToString(); val != "ssh -oBatchMode=yes -oPasswordAuthentication=no barfoo@example.com -- 'docker exec -i containerid echo '\\''foobar'\\'''" {
		t.Fatal("command builder not expected command:", val)
	}

	cmd = shell.Cmd(conn.RawCommandBuilder("echo", "foobar")...)
	if val := cmd.ToString(); val != "ssh -oBatchMode=yes -oPasswordAuthentication=no barfoo@example.com -- 'docker exec -i containerid echo foobar'" {
		t.Fatal("command builder not expected command:", val)
	}
}

