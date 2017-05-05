package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/guywithnose/runner"

	calendar "google.golang.org/api/calendar/v3"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Client gets a token for a user
type Client struct {
	config         *oauth2.Config
	tokenCacheFile string
	cmdBuilder     runner.Builder
}

// NewClient returns a Client
func NewClient(appCredentialFile, tokenCacheFile string, cmdBuilder runner.Builder) (*Client, error) {
	appCredentials, err := ioutil.ReadFile(appCredentialFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to read app credential file: %v", err)
	}

	config, err := google.ConfigFromJSON(appCredentials, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse app credentials: %v", err)
	}

	return &Client{
		config:         config,
		tokenCacheFile: tokenCacheFile,
		cmdBuilder:     cmdBuilder,
	}, nil
}

// GetHTTPClient gets an oauth token.  If necessary it may open a browser for user authorization.
func (client Client) GetHTTPClient(writer io.Writer) (*http.Client, error) {
	token, err := client.tokenFromFile()
	if err != nil {
		token, err = client.getTokenFromWeb(writer)
		if err != nil {
			return nil, err
		}

		err = client.saveToken(token)
		if err != nil {
			return nil, err
		}
	}

	return client.config.Client(context.Background(), token), err
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func (client Client) getTokenFromWeb(writer io.Writer) (*oauth2.Token, error) {
	token := make(chan string)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Your inbox should now be authorized.  You may close this window."))
		token <- r.FormValue("code")
	}))
	client.config.RedirectURL = server.URL

	authURL := client.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Fprintf(writer, "Attempting to open %s in your browser\n", authURL)
	cmd := client.cmdBuilder.New("", "xdg-open", authURL)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(writer, "Unable to open browser automatically: %v\nPlease open %s in your browser\n", err, authURL)
	}

	code := <-token
	server.Close()

	tok, err := client.config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token from web: %v", err)
	}

	return tok, nil
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func (client Client) tokenFromFile() (*oauth2.Token, error) {
	f, err := os.Open(client.tokenCacheFile)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	_ = f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func (client Client) saveToken(token *oauth2.Token) error {
	f, err := os.OpenFile(client.tokenCacheFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Unable to cache oauth token: %v", err)
	}
	err = json.NewEncoder(f).Encode(token)
	_ = f.Close()
	return err
}
