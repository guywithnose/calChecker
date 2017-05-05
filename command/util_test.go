package command_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func appWithTestWriters() (*cli.App, *bytes.Buffer) {
	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer
	return app, writer
}

func removeFile(t *testing.T, fileName string) {
	assert.Nil(t, os.RemoveAll(fileName))
}
