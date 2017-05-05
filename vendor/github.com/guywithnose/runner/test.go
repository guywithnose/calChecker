package runner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Test is used for testing code that runs system commands without actually running the commands
type Test struct {
	ExpectedCommands []*ExpectedCommand
	Errors           []error
	AnyOrder         bool
	mutex            sync.Mutex
}

// TestCommand emulates an os.exec.Cmd, but returns mock data
type TestCommand struct {
	cmdBuilder      *Test
	Dir             string
	expectedCommand *ExpectedCommand
	actualCommand   string
}

// ExpectedCommand represents a command that will be handled by a TestCommand
type ExpectedCommand struct {
	command      string
	commandRegex *regexp.Regexp
	output       []byte
	path         string
	error        error
	env          []string
	Closure      func(string)
}

// New returns a TestCommand
func (testBuilder *Test) New(path string, command ...string) Command {
	return testBuilder.validateCommand(path, nil, command...)
}

// NewWithEnvironment returns a TestCommand
func (testBuilder *Test) NewWithEnvironment(path string, env []string, command ...string) Command {
	return testBuilder.validateCommand(path, env, command...)
}

func (testBuilder *Test) validateCommand(path string, env []string, command ...string) TestCommand {
	var expectedCommand *ExpectedCommand
	commandString := strings.Join(command, " ")
	if len(testBuilder.ExpectedCommands) == 0 {
		testBuilder.Errors = append(testBuilder.Errors, fmt.Errorf("More commands were run than expected.  Extra command: %s", commandString))
	} else {
		if testBuilder.AnyOrder {
			expectedCommand = testBuilder.validateUnorderedCommand(path, commandString, env)
		} else {
			expectedCommand = testBuilder.validateOrderedCommand(path, commandString, env)
		}
	}

	return TestCommand{cmdBuilder: testBuilder, Dir: path, expectedCommand: expectedCommand, actualCommand: commandString}
}

func (testBuilder *Test) validateUnorderedCommand(path, commandString string, env []string) *ExpectedCommand {
	testBuilder.mutex.Lock()
	defer testBuilder.mutex.Unlock()
	var err error
	for index, expectedCommand := range testBuilder.ExpectedCommands {
		if expectedCommand.path != path {
			err = fmt.Errorf("Path %s did not match expected path %s", path, expectedCommand.path)
			continue
		} else if !expectedCommand.commandRegex.MatchString(commandString) {
			err = fmt.Errorf("Command '%s' did not match expected command '%s'", commandString, expectedCommand.command)
			continue
		} else if !reflect.DeepEqual(expectedCommand.env, env) {
			err = fmt.Errorf("Environment %v did not match expected environment %v", env, expectedCommand.env)
			continue
		}

		testBuilder.ExpectedCommands = append(testBuilder.ExpectedCommands[:index], testBuilder.ExpectedCommands[index+1:]...)
		return expectedCommand
	}

	testBuilder.Errors = append(testBuilder.Errors, err)
	return nil
}

func (testBuilder *Test) validateOrderedCommand(path, commandString string, env []string) *ExpectedCommand {
	expectedCommand := testBuilder.ExpectedCommands[0]
	if expectedCommand.path != path {
		testBuilder.Errors = append(testBuilder.Errors, fmt.Errorf("Path %s did not match expected path %s", path, expectedCommand.path))
	} else if !expectedCommand.commandRegex.MatchString(commandString) {
		testBuilder.Errors = append(testBuilder.Errors, fmt.Errorf("Command '%s' did not match expected command '%s'", commandString, expectedCommand.command))
	} else if !reflect.DeepEqual(expectedCommand.env, env) {
		testBuilder.Errors = append(testBuilder.Errors, fmt.Errorf("Environment %v did not match expected environment %v", env, expectedCommand.env))
	} else {
		testBuilder.ExpectedCommands = testBuilder.ExpectedCommands[1:]
	}

	return expectedCommand
}

// Output returns the expected mock data
func (cmd TestCommand) Output() ([]byte, error) {
	return cmd.run()
}

// CombinedOutput returns the expected mock data
func (cmd TestCommand) CombinedOutput() ([]byte, error) {
	return cmd.run()
}

func (cmd TestCommand) run() ([]byte, error) {
	if cmd.expectedCommand == nil {
		return nil, nil
	}

	if cmd.expectedCommand.Closure != nil {
		cmd.expectedCommand.Closure(cmd.actualCommand)
	}

	return cmd.expectedCommand.output, cmd.expectedCommand.error
}

// NewExpectedCommand returns a new ExpectedCommand
func NewExpectedCommand(path, command, output string, exitCode int) *ExpectedCommand {
	commandRegex, err := regexp.Compile(fmt.Sprintf("^%s$", command))
	if err != nil {
		panic(err)
	}

	if exitCode == -1 {
		err = errors.New("Error running command")
	} else if exitCode != 0 {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--", strconv.Itoa(exitCode))
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		err = cmd.Run()
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitErr.Stderr = []byte(output)
			err = exitErr
		}
	}

	return &ExpectedCommand{
		commandRegex: commandRegex,
		command:      command,
		output:       []byte(output),
		path:         path,
		error:        err,
	}
}

// WithEnvironment adds environment expectations to an ExpectedCommand
func (ec *ExpectedCommand) WithEnvironment(env []string) *ExpectedCommand {
	ec.env = env
	return ec
}

// ErrorCodeHelper exits with a specified error code
// This is used in tests that require a command to return an error code other than 0
// To use this the test must include a test like this:
//
// func TestHelperProcess(*testing.T) {
//     commandBuilder.ErrorCodeHelper()
// }
func ErrorCodeHelper() {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	code, err := strconv.Atoi(os.Args[3])
	if err != nil {
		code = 1
	}

	defer os.Exit(code)
}
