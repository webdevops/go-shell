package shell

func CommandInterfaceBuilder(command string, args ...string) []interface{} {
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

	return shellCmd
}
