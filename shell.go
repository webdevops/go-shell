package shell

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	ShellList   = map[string][]string{
		"bash": {"/bin/bash",  "-o", "errexit", "-o", "pipefail", "-c"},
		"sh":   {"/bin/sh",  "-o", "errexit", "-c"},
	}
	Shell       = []string{"/bin/sh",  "-o", "errexit", "-c"}
	Panic       = true
	ErrorFunc   = func(c *Command, p *Process) {}
	VerboseFunc = func(c *Command) {}
	Trace       = false
	TracePrefix = "+"

	exit = os.Exit
)

var Tee io.Writer

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

func SetDefaultShell(shell string) {
	if val, ok := ShellList[shell]; ok {
		Shell = val
	} else {
		panic(fmt.Sprintf("Shell %v is not supported"))
	}
}

func NewCmd(command string, args ...string) *Command {
	cmd := []string{
		command,
	}

	for _, val := range args {
		cmd = append(cmd, val)
	}

	shellCmd := make([]interface{}, len(cmd))
	for i, v := range cmd {
		shellCmd[i] = v
	}

	return Cmd(shellCmd...)
}

func Path(parts ...string) string {
	return filepath.Join(parts...)
}

func PathTemplate(parts ...string) func(...interface{}) string {
	return func(values ...interface{}) string {
		return fmt.Sprintf(Path(parts...), values...)
	}
}

func Quote(arg string) string {
	return fmt.Sprintf("'%s'", strings.Replace(arg, "'", "'\\''", -1))
}

func QuoteValues(arg ...string) []string {
	for i, v := range arg {
		arg[i] = Quote(v)
	}
	return arg
}

func ErrExit() {
	if p, ok := recover().(*Process); p != nil {
		if !ok {
			fmt.Fprintf(os.Stderr, "Unexpected panic: %v\n", p)
			exit(1)
		}
		fmt.Fprintf(os.Stderr, "%s\n", p.Error())
		exit(p.ExitStatus)
	}
}

type Command struct {
	args []string
	in   *Command
}

func (c *Command) ProcFn() func(...interface{}) *Process {
	return func(args ...interface{}) *Process {
		cmd := &Command{c.args, c.in}
		cmd.addArgs(args...)
		return cmd.Run()
	}
}

func (c *Command) OutputFn() func(...interface{}) (string, error) {
	return func(args ...interface{}) (out string, err error) {
		cmd := &Command{c.args, c.in}
		cmd.addArgs(args...)
		defer func() {
			if p, ok := recover().(*Process); p != nil {
				if ok {
					err = p.Error()
				} else {
					err = fmt.Errorf("panic: %v", p)
				}
			}
		}()
		out = cmd.Run().String()
		return
	}
}

func (c *Command) ErrFn() func(...interface{}) error {
	return func(args ...interface{}) (err error) {
		cmd := &Command{c.args, c.in}
		cmd.addArgs(args...)
		defer func() {
			if p, ok := recover().(*Process); p != nil {
				if ok {
					err = p.Error()
				} else {
					err = fmt.Errorf("panic: %v", p)
				}
			}
		}()
		cmd.Run()
		return
	}
}

func (c *Command) Pipe(cmd ...interface{}) *Command {
	return Cmd(append(cmd, c)...)
}

func (c *Command) addArgs(args ...interface{}) {
	var strArgs []string
	for i, arg := range args {
		switch v := arg.(type) {
		case string:
			strArgs = append(strArgs, v)
		case fmt.Stringer:
			strArgs = append(strArgs, v.String())
		default:
			cmd, ok := arg.(*Command)
			if i+1 == len(args) && ok {
				c.in = cmd
				continue
			}
			panic("invalid type for argument")
		}
	}
	c.args = append(c.args, strArgs...)
}

func (c *Command) shellCmd(quote bool) string {
	if !quote {
		return strings.Join(c.args, " ")
	}
	var quoted []string
	for i := range c.args {
		quoted = append(quoted, Quote(c.args[i]))
	}
	return strings.Join(quoted, " ")
}

func (c *Command) ToString() string {
	var ret []string

	if c.in != nil {
		ret = append(ret, c.in.ToString())
	}

	ret = append(ret, c.shellCmd(false))

	return strings.Join(ret, " | ")
}

func (c *Command) Run() *Process {
	VerboseFunc(c)
	return c.execute(false)
}

func (c *Command) RunInteractive() *Process {
	VerboseFunc(c)
	return c.execute(true)
}

func (c *Command) execute(interactive bool) *Process {
	if Trace {
		fmt.Fprintln(os.Stderr, TracePrefix, c.shellCmd(false))
	}
	cmd := exec.Command(Shell[0], append(Shell[1:], c.shellCmd(false))...)
	p := new(Process)
	p.Command = c
	if c.in != nil {
		cmd.Stdin = c.in.execute(false)
	} else {
		stdin, err := cmd.StdinPipe()
		assert(err)
		p.Stdin = stdin
	}

	if interactive {
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
	} else {
		var stdout bytes.Buffer
		if Tee != nil {
			cmd.Stdout = io.MultiWriter(&stdout, Tee)
		} else {
			cmd.Stdout = &stdout
		}
		p.Stdout = &stdout
		var stderr bytes.Buffer
		if Tee != nil {
			cmd.Stderr = io.MultiWriter(&stderr, Tee)
		} else {
			cmd.Stderr = &stderr
		}
		p.Stderr = &stderr
	}
	err := cmd.Run()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if stat, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				p.ExitStatus = int(stat.ExitStatus())
				ErrorFunc(c, p)
				if Panic {
					panic(p)
				}
			}
		} else {
			assert(err)
		}
	}
	return p
}

func Cmd(cmd ...interface{}) *Command {
	c := new(Command)
	c.addArgs(cmd...)
	return c
}

type Process struct {
	Stdout     *bytes.Buffer
	Stderr     *bytes.Buffer
	Stdin      io.WriteCloser
	ExitStatus int
	Command    *Command
}

func (p *Process) String() string {
	stderr := strings.Replace(p.Stderr.String(), "\n", "\n           ", -1)

	msg := "go-shell command failed\n"
	msg += fmt.Sprintf("COMMAND:   %v\n", p.Command.ToString())
	msg += fmt.Sprintf("EXIT CODE: %v\n", p.ExitStatus)
	msg += fmt.Sprintf("STDERR:    %v\n", stderr)
	msg += "\n"

	return msg
}

func (p *Process) Bytes() []byte {
	return p.Stdout.Bytes()
}

func (p *Process) Error() error {
	errlines := strings.Split(p.Stderr.String(), "\n")
	return fmt.Errorf("[%v] %s\n", p.ExitStatus, errlines[len(errlines)-2])
}

func (p *Process) Read(b []byte) (int, error) {
	return p.Stdout.Read(b)
}

func (p *Process) Write(b []byte) (int, error) {
	return p.Stdin.Write(b)
}

func Run(cmd ...interface{}) *Process {
	return Cmd(cmd...).Run()
}
