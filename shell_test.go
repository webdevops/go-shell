package shell

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestCmdRun(t *testing.T) {
	output := Cmd("echo", "foobar").Run().String()
	if output != "foobar" {
		t.Fatal("output not expected:", output)
	}
}

func TestRun(t *testing.T) {
	output := Run("echo", "foobar").String()
	if output != "foobar" {
		t.Fatal("output not expected:", output)
	}
}

func TestPanic(t *testing.T) {
	defer func() {
		p := recover().(*Process).ExitStatus
		if p != 2 {
			t.Fatal("status not expected:", p)
		}
	}()
	Run("exit", "2")
}

func TestPipe(t *testing.T) {
	p := Cmd("echo", "foobar").Pipe("wc", "-c").Pipe("awk", "'{print $1}'").Run()
	if p.String() != "7" {
		t.Fatal("output not expected:", p.String())
	}
}

func TestSingleArg(t *testing.T) {
	p := Run("echo foobar | wc -c | awk '{print $1}'")
	if p.String() != "7" {
		t.Fatal("output not expected:", p.String())
	}
}

func TestProcessAsArg(t *testing.T) {
	p := Cmd("echo", Run("echo foobar")).Pipe("wc", "-c").Pipe(Run("echo", "awk"), "'{print $1}'").Run()
	if p.String() != "7" {
		t.Fatal("output not expected:", p.String())
	}
}

func TestLastArgStdin(t *testing.T) {
	p := Cmd("awk '{print $1}'", Cmd("wc", "-c", Cmd("echo foobar"))).Run()
	if p.String() != "7" {
		t.Fatal("output not expected:", p.String())
	}
}

func TestCmdRunFunc(t *testing.T) {
	echo := Cmd("echo").ProcFn()
	output := echo("foobar").String()
	if output != "foobar" {
		t.Fatal("output not expected:", output)
	}
}

func TestPath(t *testing.T) {
	p := Path("/root", "part1/part2", "foobar")
	if p != "/root/part1/part2/foobar" {
		t.Fatal("path not expected:", p)
	}
}

func TestPathTemplate(t *testing.T) {
	tmpl := PathTemplate("/root", "%s/part2", "%s")
	p := tmpl("one", "two")
	if p != "/root/one/part2/two" {
		t.Fatal("path not expected:", p)
	}
}

func TestPrintlnStringer(t *testing.T) {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, Run("echo foobar"))
	if buf.String() != "foobar\n" {
		t.Fatal("output not expected:", buf.String())
	}
}

func TestWrapPanicToErr(t *testing.T) {
	copy := func(src, dst string) (err error) {
		defer func() {
			if p := recover().(*Process); p != nil {
				err = p.Error()
			}
		}()
		Run("cp", src, dst)
		return
	}

	err := copy("", "")
	if !(strings.HasPrefix(err.Error(), "[64] ") || strings.HasPrefix(err.Error(), "[1] ")) {
		// osx and linux error is expected
		t.Fatal("output not expected:", err)
	}
}

func TestCmdOutputFn(t *testing.T) {
	copy := Cmd("cp").OutputFn()
	echo := Cmd("echo").OutputFn()
	_, err := copy("", "")
	if !(strings.HasPrefix(err.Error(), "[64] ") || strings.HasPrefix(err.Error(), "[1] ")) {
		t.Fatal("output not expected:", err)
	}
	out, err := echo("foobar")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if out != "foobar" {
		t.Fatal("output not expected:", out)
	}
}

func TestToString(t *testing.T) {
	var command string

	command = Cmd("echo foobar").Pipe("wc -c").ToString()
	if command != "echo foobar | wc -c" {
		t.Fatal("output not expected:", command)
	}

	command = Cmd("echo foobar").Pipe("wc -c").Pipe("wc -l").ToString()
	if command != "echo foobar | wc -c | wc -l" {
		t.Fatal("output not expected:", command)
	}
}
