package command_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guywithnose/calChecker/command"
	"github.com/guywithnose/runner"
	"github.com/stretchr/testify/assert"
)

func TestGetHTTPCLient(t *testing.T) {
	testFolder := filepath.Join(os.TempDir(), "testCalChecker")
	defer removeFile(t, testFolder)
	assert.Nil(t, os.MkdirAll(testFolder, 0777))
	credentialFile := filepath.Join(testFolder, "credentials")
	tokenCacheFile := filepath.Join(testFolder, "token")
	ts := getMockGoogleAPI(t)
	defer ts.Close()
	assert.Nil(t, ioutil.WriteFile(credentialFile, getTestCredentials(ts.URL), 0777))
	ec := runner.NewExpectedCommand("", "xdg-open.*", "", 0)
	var OAuthURL string
	ec.Closure = func(command string) {
		OAuthURL = strings.Replace(command, "xdg-open ", "", -1)
		_, err := http.Get(OAuthURL)
		assert.Nil(t, err)
	}
	cb := &runner.Test{ExpectedCommands: []*runner.ExpectedCommand{ec}}
	client, err := command.NewClient(credentialFile, tokenCacheFile, cb)
	assert.Nil(t, err)
	assert.NotNil(t, client)
	writer := &bytes.Buffer{}
	httpClient, err := client.GetHTTPClient(writer)
	assert.Nil(t, err)
	assert.NotNil(t, httpClient)
	assert.Equal(t, []*runner.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	tokenFileContents, _ := ioutil.ReadFile(tokenCacheFile)
	assert.Equal(t, "{\"access_token\":\"fakeToken\",\"expiry\":\"0001-01-01T00:00:00Z\"}\n", string(tokenFileContents))
	assert.Equal(t, fmt.Sprintf("Attempting to open %s in your browser\n", OAuthURL), writer.String())
}

func TestGetHTTPCLientTokenFailure(t *testing.T) {
	testFolder := filepath.Join(os.TempDir(), "testCalChecker")
	defer removeFile(t, testFolder)
	assert.Nil(t, os.MkdirAll(testFolder, 0777))
	credentialFile := filepath.Join(testFolder, "credentials")
	tokenCacheFile := filepath.Join(testFolder, "token")
	ts := getMockGoogleAPITokenFailure(t)
	defer ts.Close()
	assert.Nil(t, ioutil.WriteFile(credentialFile, getTestCredentials(ts.URL), 0777))
	ec := runner.NewExpectedCommand("", "xdg-open.*", "", 0)
	ec.Closure = func(command string) {
		_, err := http.Get(strings.Replace(command, "xdg-open ", "", -1))
		assert.Nil(t, err)
	}
	cb := &runner.Test{ExpectedCommands: []*runner.ExpectedCommand{ec}}
	client, err := command.NewClient(credentialFile, tokenCacheFile, cb)
	assert.Nil(t, err)
	assert.NotNil(t, client)
	writer := &bytes.Buffer{}
	httpClient, err := client.GetHTTPClient(writer)
	assert.EqualError(t, err, "Unable to retrieve token from web: oauth2: cannot fetch token: 500 Internal Server Error\nResponse: ")
	assert.Nil(t, httpClient)
	assert.Equal(t, []*runner.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestGetHTTPCLientInvalidTokenFile(t *testing.T) {
	testFolder := filepath.Join(os.TempDir(), "testCalChecker")
	defer removeFile(t, testFolder)
	assert.Nil(t, os.MkdirAll(testFolder, 0777))
	credentialFile := filepath.Join(testFolder, "credentials")
	tokenCacheFile := filepath.Join(testFolder, "doesntexist", "token")
	ts := getMockGoogleAPI(t)
	defer ts.Close()
	assert.Nil(t, ioutil.WriteFile(credentialFile, getTestCredentials(ts.URL), 0777))
	ec := runner.NewExpectedCommand("", "xdg-open.*", "", 0)
	ec.Closure = func(command string) {
		_, err := http.Get(strings.Replace(command, "xdg-open ", "", -1))
		assert.Nil(t, err)
	}
	cb := &runner.Test{ExpectedCommands: []*runner.ExpectedCommand{ec}}
	client, err := command.NewClient(credentialFile, tokenCacheFile, cb)
	assert.Nil(t, err)
	assert.NotNil(t, client)
	writer := &bytes.Buffer{}
	httpClient, err := client.GetHTTPClient(writer)
	assert.EqualError(t, err, "Unable to cache oauth token: open /tmp/testCalChecker/doesntexist/token: no such file or directory")
	assert.Nil(t, httpClient)
	assert.Equal(t, []*runner.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestGetHTTPCLientLoadFromCacheFile(t *testing.T) {
	testFolder := filepath.Join(os.TempDir(), "testCalChecker")
	defer removeFile(t, testFolder)
	assert.Nil(t, os.MkdirAll(testFolder, 0777))
	credentialFile := filepath.Join(testFolder, "credentials")
	tokenCacheFile := filepath.Join(testFolder, "token")
	assert.Nil(t, ioutil.WriteFile(tokenCacheFile, []byte("{\"access_token\":\"fakeToken\",\"expiry\":\"0001-01-01T00:00:00Z\"}\n"), 0777))
	ts := getMockGoogleAPI(t)
	defer ts.Close()
	assert.Nil(t, ioutil.WriteFile(credentialFile, getTestCredentials(ts.URL), 0777))
	cb := &runner.Test{}
	client, err := command.NewClient(credentialFile, tokenCacheFile, cb)
	assert.Nil(t, err)
	assert.NotNil(t, client)
	writer := &bytes.Buffer{}
	httpClient, err := client.GetHTTPClient(writer)
	assert.Nil(t, err)
	assert.NotNil(t, httpClient)
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestGetHTTPCLientUnableToOpenBrowser(t *testing.T) {
	testFolder := filepath.Join(os.TempDir(), "testCalChecker")
	defer removeFile(t, testFolder)
	assert.Nil(t, os.MkdirAll(testFolder, 0777))
	credentialFile := filepath.Join(testFolder, "credentials")
	tokenCacheFile := filepath.Join(testFolder, "token")
	ts := getMockGoogleAPI(t)
	defer ts.Close()
	assert.Nil(t, ioutil.WriteFile(credentialFile, getTestCredentials(ts.URL), 0777))
	ec := runner.NewExpectedCommand("", "xdg-open.*", "", 1)
	var OAuthURL string
	ec.Closure = func(command string) {
		OAuthURL = strings.Replace(command, "xdg-open ", "", -1)
		_, err := http.Get(OAuthURL)
		assert.Nil(t, err)
	}
	cb := &runner.Test{ExpectedCommands: []*runner.ExpectedCommand{ec}}
	client, err := command.NewClient(credentialFile, tokenCacheFile, cb)
	assert.Nil(t, err)
	assert.NotNil(t, client)
	writer := &bytes.Buffer{}
	httpClient, err := client.GetHTTPClient(writer)
	assert.Nil(t, err)
	assert.NotNil(t, httpClient)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(
		t,
		[]string{
			fmt.Sprintf("Attempting to open %s in your browser", OAuthURL),
			"Unable to open browser automatically: exit status 1",
			fmt.Sprintf("Please open %s in your browser", OAuthURL),
			"",
		},
		strings.Split(writer.String(), "\n"),
	)
}

func TestGetHTTPCLientInvalidCredentialFile(t *testing.T) {
	testFolder := filepath.Join(os.TempDir(), "testCalChecker")
	defer removeFile(t, testFolder)
	assert.Nil(t, os.MkdirAll(testFolder, 0777))
	credentialFile := filepath.Join(testFolder, "credentials")
	tokenCacheFile := filepath.Join(testFolder, "token")
	cb := &runner.Test{}
	_, err := command.NewClient(credentialFile, tokenCacheFile, cb)
	assert.EqualError(t, err, "Unable to read app credential file: open /tmp/testCalChecker/credentials: no such file or directory")
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestGetHTTPCLientCorruptCredentialFile(t *testing.T) {
	testFolder := filepath.Join(os.TempDir(), "testCalChecker")
	defer removeFile(t, testFolder)
	assert.Nil(t, os.MkdirAll(testFolder, 0777))
	credentialFile := filepath.Join(testFolder, "credentials")
	assert.Nil(t, ioutil.WriteFile(credentialFile, []byte("not json"), 0777))
	tokenCacheFile := filepath.Join(testFolder, "token")
	cb := &runner.Test{}
	_, err := command.NewClient(credentialFile, tokenCacheFile, cb)
	assert.EqualError(t, err, "Unable to parse app credentials: invalid character 'o' in literal null (expecting 'u')")
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestHelperProcess(*testing.T) {
	runner.ErrorCodeHelper()
}
