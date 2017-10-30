package commandbuilder

import (
	"testing"
)

func TestArgumentPlain(t *testing.T) {
	value := "foobar"
	argument, err := ParseArgument(value)

	if err != nil {
		t.Fatal("Error thrown:", err)
	}

	if argument.Raw != value {
		t.Fatal("argument value not expected:", argument.Raw)
	}
}

func TestArgumentDsnStyle(t *testing.T) {
	value := "docker:mycontainer;path=/my/custom/path;env[foo]=BAR;user=barfoo"
	argument, err := ParseArgument(value)

	if err != nil {
		t.Fatal("Error thrown:", err)
	}

	if argument.Raw != value {
		t.Fatal("argument value not expected:", argument.Raw)
	}

	if argument.Scheme != "docker" {
		t.Fatal("argument scheme not expected:", argument.Scheme)
	}

	if argument.Hostname != "mycontainer" {
		t.Fatal("argument hostname not expected:", argument.Hostname)
	}

	if argument.Username != "barfoo" {
		t.Fatal("argument username not expected:", argument.Username)
	}

	if arg, err := argument.GetOption("path"); err == nil {
		if arg != "/my/custom/path" {
			t.Fatal("argument option path not expected:", arg)
		}
	} else {
		t.Fatal("Error thrown:", err)
	}

	if env, err := argument.GetEnvironment("foo"); err == nil {
		if env != "BAR" {
			t.Fatal("argument environment foo not expected:", env)
		}
	} else {
		t.Fatal("Error thrown:", err)
	}
}


func TestArgumentDsnStyleHostAsParam(t *testing.T) {
	value := "docker:host=mycontainer;path=/my/custom/path;env[foo]=BAR;user=barfoo"
	argument, err := ParseArgument(value)

	if err != nil {
		t.Fatal("Error thrown:", err)
	}

	if argument.Raw != value {
		t.Fatal("argument value not expected:", argument.Raw)
	}

	if argument.Scheme != "docker" {
		t.Fatal("argument scheme not expected:", argument.Scheme)
	}

	if argument.Hostname != "mycontainer" {
		t.Fatal("argument hostname not expected:", argument.Hostname)
	}

	if argument.Username != "barfoo" {
		t.Fatal("argument username not expected:", argument.Username)
	}

	if arg, err := argument.GetOption("path"); err == nil {
		if arg != "/my/custom/path" {
			t.Fatal("argument option path not expected:", arg)
		}
	} else {
		t.Fatal("Error thrown:", err)
	}

	if env, err := argument.GetEnvironment("foo"); err == nil {
		if env != "BAR" {
			t.Fatal("argument environment foo not expected:", env)
		}
	} else {
		t.Fatal("Error thrown:", err)
	}
}

func TestArgumentUrlStyle(t *testing.T) {
	value := "docker://mycontainer?path=/my/custom/path&env[foo]=BAR"
	argument, err := ParseArgument(value)

	if err != nil {
		t.Fatal("Error thrown:", err)
	}

	if argument.Raw != value {
		t.Fatal("argument raw not expected:", argument.Raw)
	}

	if argument.Scheme != "docker" {
		t.Fatal("argument scheme not expected:", argument.Scheme)
	}

	if argument.Hostname != "mycontainer" {
		t.Fatal("argument hostname not expected:", argument.Hostname)
	}

	if arg, err := argument.GetOption("path"); err == nil {
		if arg != "/my/custom/path" {
			t.Fatal("argument option path not expected:", arg)
		}
	} else {
		t.Fatal("Error thrown:", err)
	}

	if env, err := argument.GetEnvironment("foo"); err == nil {
		if env != "BAR" {
			t.Fatal("argument environment foo not expected:", env)
		}
	} else {
		t.Fatal("Error thrown:", err)
	}
}
