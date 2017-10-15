package commandbuilder

import (
	"strings"
	"regexp"
	"net/url"
	"fmt"
	"errors"
)

var argParserDsn = regexp.MustCompile("^(?P<schema>[-_a-zA-Z0-9]+):(?P<container>[^/;]+)(?P<params>;[^/;]+.*)?$")
var argParserUrl = regexp.MustCompile("^[-_a-zA-Z0-9]+://[^/]+.*$")
var argParserEnv = regexp.MustCompile("^env\\[(?P<item>[a-zA-Z0-9]+)\\]$")

type Argument struct {
	url.URL
	Value string
	Options map[string]string
	Environment map[string]string
	simple bool
}

func ParseArgument(value string) (Argument, error) {
	argument := Argument{}
	argument.Options = map[string]string{}
	argument.Environment = map[string]string{}
	argument.Value = value
	argument.simple = true

	if argParserDsn.MatchString(value) {
		err := argument.ParseDsn()
		if err != nil {
			return argument, err
		}
		argument.simple = false
	} else if argParserUrl.MatchString(value) {
		err := argument.ParseUrl()
		if err != nil {
			return argument, err
		}
		argument.simple = false
	}

	return argument, nil
}

func (argument *Argument) ParseDsn() error {
	match := argParserDsn.FindStringSubmatch(argument.Value)
	argument.Scheme = match[1]
	argument.Host = match[2]

	paramString := match[3]
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
				// Default option
				argument.Options[varName] = varValue
			}
		}
	}


	// Init user
	if argument.HasOption("username") && argument.HasOption("password") {
		username, _ := argument.GetOption("username")
		password, _ := argument.GetOption("password")
		argument.User = url.UserPassword(username, password)
	} else if argument.HasOption("username") {
		username, _ := argument.GetOption("username")
		argument.User = url.User(username)
	}

	return nil
}

func (argument *Argument) ParseUrl() error {
	// parse url
	parsedUrl, err := url.Parse(argument.Value)
	if err != nil {
		return err
	}

	// copy parsed url informations into argument.URL
	argument.URL = *parsedUrl

	for varName, varValue := range parsedUrl.Query() {
		if argParserEnv.MatchString(varName) {
			// env value
			match := argParserEnv.FindStringSubmatch(varName)
			envName := match[1]
			argument.Environment[envName] = varValue[0]
		} else {
			// Default option
			argument.Options[varName] = varValue[0]
		}
	}

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
