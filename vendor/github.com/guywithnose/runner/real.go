package runner

import "os/exec"

// Real is the standard CommandBuilder that will return instances of os.exec.Cmd
type Real struct{}

// New creates an instance os.exec.Cmd
func (w Real) New(path string, command ...string) Command {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = path
	return cmd
}

// NewWithEnvironment creates an instance os.exec.Cmd
func (w Real) NewWithEnvironment(path string, env []string, command ...string) Command {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = path
	cmd.Env = env
	return cmd
}
