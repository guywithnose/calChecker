package runner_test

import (
	"os/exec"
	"testing"

	"github.com/guywithnose/runner"
	"github.com/stretchr/testify/assert"
)

func TestRealRunner(t *testing.T) {
	cb := &runner.Real{}
	command := cb.New("", "ls")
	assert.IsType(t, &exec.Cmd{}, command)
}

func TestRealRunnerImplementsBuilder(t *testing.T) {
	var _ runner.Builder = (*runner.Real)(nil)
}

func TestRealRunnerWithEnvironment(t *testing.T) {
	cb := &runner.Real{}
	command := cb.NewWithEnvironment("", []string{"FOO=BAR"}, "ls")
	assert.IsType(t, &exec.Cmd{}, command)
	cmd := command.(*exec.Cmd)
	assert.Equal(t, []string{"FOO=BAR"}, cmd.Env)
}
