package command

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/guywithnose/runner"
	"github.com/urfave/cli"
	calendar "google.golang.org/api/calendar/v3"
)

// BasePath allows overriding the calendar API base path for testing
var BasePath string

// CmdCheck checks the inbox for unread messages
func CmdCheck(cmdBuilder runner.Builder) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if c.NArg() != 0 {
			return cli.NewExitError("Usage: \"calChecker\"", 1)
		}

		err := checkFlags(c)
		if err != nil {
			return err
		}

		srv, err := getCalendarService(c.String("credentialFile"), c.String("tokenFile"), c.App.Writer, cmdBuilder)
		if err != nil {
			return err
		}

		request := srv.CalendarList.List()
		resp, err := request.Do()
		if err != nil {
			return fmt.Errorf("Unable to check calendar. %v", err)
		}

		err = parseCalendars(srv, resp.Items, c.App.Writer)
		if err != nil {
			return err
		}

		for resp.NextPageToken != "" {
			request.PageToken(resp.NextPageToken)
			resp, err = request.Do()
			if err != nil {
				return fmt.Errorf("Unable to check calendar. %v", err)
			}

			err = parseCalendars(srv, resp.Items, c.App.Writer)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func getCalendarService(credentialFile, tokenFile string, w io.Writer, cmdBuilder runner.Builder) (*calendar.Service, error) {
	tokenClient, err := NewClient(credentialFile, tokenFile, cmdBuilder)
	if err != nil {
		return nil, fmt.Errorf("Could not initialize token client: %v", err)
	}

	httpClient, err := tokenClient.GetHTTPClient(w)
	if err != nil {
		return nil, fmt.Errorf("Could not get OAuth token: %v", err)
	}

	srv, _ := calendar.New(httpClient)
	if BasePath != "" {
		srv.BasePath = BasePath
	}

	return srv, nil
}

func parseCalendars(srv *calendar.Service, items []*calendar.CalendarListEntry, w io.Writer) error {
	for _, item := range items {
		if item.Primary {
			midnight := fmt.Sprintf("%sT00:00:00Z", time.Now().Format("2006-01-02"))
			tomorrow := fmt.Sprintf("%sT00:00:00Z", time.Now().Add(time.Hour*24).Format("2006-01-02"))
			request := srv.Events.List(item.Id).TimeMin(midnight).TimeMax(tomorrow).SingleEvents(true)
			resp, err := request.Do()
			if err != nil {
				return fmt.Errorf("Unable to check calendar. %v", err)
			}

			err = parseEvents(resp.Items, w)
			if err != nil {
				return err
			}

			for resp.NextPageToken != "" {
				request.PageToken(resp.NextPageToken)
				resp, err = request.Do()
				if err != nil {
					return err
				}

				err = parseEvents(resp.Items, w)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func parseEvents(items []*calendar.Event, w io.Writer) error {
	tabW := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tabW.Flush()
	for _, event := range items {
		if event.Start.DateTime != "" {
			start, err := time.Parse(time.RFC3339, event.Start.DateTime)
			if err != nil {
				return err
			}

			fmt.Fprintf(tabW, "%s\t%s\n", start.Format("Mon, 3:04PM"), event.Summary)
		} else {
			fmt.Fprintf(tabW, "All Day\t%s\n", event.Summary)
		}
	}

	return nil
}

func checkFlags(c *cli.Context) error {
	if c.String("credentialFile") == "" {
		return cli.NewExitError("You must specify a credentialFile", 1)
	}

	if c.String("tokenFile") == "" {
		return cli.NewExitError("You must specify a tokenFile", 1)
	}

	return nil
}
