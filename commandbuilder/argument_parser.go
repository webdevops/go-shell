package commandbuilder

import (
	"strings"
	"regexp"
	"net/url"
	"fmt"
	"errors"
	"github.com/mohae/deepcopy"
)

var argParserDsn = regexp.MustCompile("^(?P<schema>[-_a-zA-Z0-9]+):((?P<hostname>[^/;=]+)(?P<param1>;[^/;]+.*)*|(?P<param2>[^=]+=[^;]+)(?P<param3>;[^/;]+.*)*)$")
var argParserUrl = regexp.MustCompile("^[-_a-zA-Z0-9]+://[^/]+.*$")
var argParserUserHost = regexp.MustCompile("^(?P<user>[^@]+)@(?P<hostname>.+)$")
var argParserEnv = regexp.MustCompile("^env\\[(?P<item>[a-zA-Z0-9]+)\\]$")

type Argument struct {
	Scheme string
	Hostname string
	Port string
	Username string
	Password string
	Raw string
	Options map[string]string
	Environment map[string]string
	Workdir string
	simple bool
}

func ParseArgument(value string) (Argument, error) {
	argument := Argument{}
	err := argument.Set(value)
	return argument, err
}


func (argument *Argument) Clear() {
	argument.Scheme = ""
	argument.Hostname = ""
	argument.Port = ""
	argument.Username = ""
	argument.Password = ""
	argument.Raw = ""
	argument.Options = map[string]string{}
	argument.Environment = map[string]string{}
	argument.Workdir = ""
}

func (argument *Argument) Set(value string) error {
	argument.Clear()

	argument.Options = map[string]string{}
	argument.Environment = map[string]string{}
	argument.Raw = value
	argument.simple = true

	if argParserUrl.MatchString(value) {
		err := argument.ParseUrl()
		if err != nil {
			return err
		}
		argument.simple = false
	} else if argParserDsn.MatchString(value) {
		err := argument.ParseDsn()
		if err != nil {
			return err
		}
		argument.simple = false
	} else if argParserUserHost.MatchString(value) {
		err := argument.ParseUserHost()
		if err != nil {
			return err
		}
		argument.simple = false
	} else {
		argument.Hostname = value
		argument.simple = true
	}

	return nil
}
// Clone connection
func (argument *Argument) Clone() (*Argument) {
	clone := deepcopy.Copy(argument).(*Argument)

	if clone.Options == nil {
		clone.Options = map[string]string{}
	}

	if clone.Environment == nil {
		clone.Environment = map[string]string{}
	}

	return clone
}

func (argument *Argument) IsEmpty() bool {
	if argument.Scheme != "" { return false }
	if argument.Hostname != "" { return false }
	if argument.Port != "" { return false }
	if argument.Username != "" { return false }
	if argument.Password != "" { return false }
	if argument.Raw != "" { return false }
	if len(argument.Options) >= 1 { return false }
	if len(argument.Environment) >= 1 { return false }
	if argument.Workdir != "" { return false }
	return true
}

func (argument *Argument) ParseDsn() error {
	// parse string with regex
	match := argParserDsn.FindStringSubmatch(argument.Raw)
	result := make(map[string]string)
	for i, name := range argParserDsn.SubexpNames() {
		if i != 0 { result[name] = match[i] }
	}

	// assign default result
	argument.Scheme = result["schema"]
	argument.Hostname = result["hostname"]

	// build param string and parse it
	paramString := result["param1"] + ";" + result["param2"] + ";" + result["param3"]
	paramString = strings.Trim(paramString, ";")
	paramList := strings.Split(paramString, ";")
	for _, val := range paramList {
		split := strings.SplitN(val, "=", 2)
		if len(split) == 2 {
			varName := split[0]
			varValue := split[1]

			if argParserEnv.MatchString(varName) {
				// env value
				match := argParserEnv.FindStringSubmatch(varName)
				envName := match[1]
				argument.Environment[envName] = varValue
			} else {
				switch varName {
				case "host":
					fallthrough
				case "hostname":
					argument.Hostname = varValue
				case "port":
					argument.Port = varValue
				case "user":
					fallthrough
				case "username":
					argument.Username = varValue
				case "password":
					argument.Password = varValue
				case "workdir":
					argument.Workdir = varValue
				default:
					argument.Options[varName] = varValue
				}
			}
		}
	}

	return nil
}

func (argument *Argument) ParseUrl() error {
	// parse url
	url, err := url.Parse(argument.Raw)
	if err != nil {
		return err
	}

	// copy parsed url informations
	argument.Scheme = url.Scheme
	argument.Hostname = url.Hostname()
	argument.Port = url.Port()

	if url.User != nil && url.User.Username() != "" {
		argument.Username = url.User.Username()

		if val, ok := url.User.Password(); ok {
			argument.Password = val
		}
	}

	for varName, varValue := range url.Query() {
		if argParserEnv.MatchString(varName) {
			// env value
			match := argParserEnv.FindStringSubmatch(varName)
			envName := match[1]
			argument.Environment[envName] = varValue[0]
		} else {
			switch varName {
			case "workdir":
				argument.Workdir = varValue[0]
			default:
				argument.Options[varName] = varValue[0]
			}
		}
	}

	return nil
}


func (argument *Argument) ParseUserHost() error {
	// parse string with regex
	match := argParserUserHost.FindStringSubmatch(argument.Raw)
	result := make(map[string]string)
	for i, name := range argParserUserHost.SubexpNames() {
		if i != 0 { result[name] = match[i] }
	}

	// assign default result
	argument.Username = result["user"]
	argument.Hostname = result["hostname"]

	return nil
}

func (argument *Argument) IsSimpleValue() bool {
	return argument.simple
}

func (argument *Argument) HasOption(name string) bool {
	if _, ok := argument.Options[name]; ok {
		return true
	}
	return false
}

func (argument *Argument) GetOption(name string)  (string, error) {
	if val, ok := argument.Options[name]; ok {
		return val, nil
	} else {
		return "", errors.New(fmt.Sprintf("No option variable %s found", name))
	}
}

func (argument *Argument) HasEnvironment(name string) bool {
	if _, ok := argument.Environment[name]; ok {
		return true
	}
	return false
}

func (argument *Argument) GetEnvironment(name string) (string, error) {
	if val, ok := argument.Environment[name]; ok {
		return val, nil
	} else {
		return "", errors.New(fmt.Sprintf("No environment variable %s found", name))
	}
}
