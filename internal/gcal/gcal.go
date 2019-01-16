package gcal

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// CalendarItem is an output DTO
type CalendarItem struct {
	Start time.Time
	End   time.Time
}

// Calendar is an output DTO
type Calendar struct {
	Items []*CalendarItem
}

// CalendarAPI contains the functions to call the calendar APIs
type CalendarAPI struct{}

// GetCalendars returns the calendars for the emails (map values) provided
func (c *CalendarAPI) GetCalendars(credentialsFile, tokenFile string, users map[string]string, start time.Time, end time.Time) (map[string]*Calendar, error) {
	api, err := c.getAPI(credentialsFile, tokenFile)
	if err != nil {
		return nil, err
	}

	out := map[string]*Calendar{}

	for id, email := range users {
		calendar, err := c.getCalendar(api, email, start, end)
		if err != nil {
			return nil, err
		}

		out[id] = calendar
	}

	return out, nil
}

// will return the calendar for the supplied email address
// (taken from API example)
func (c *CalendarAPI) getCalendar(api *calendar.Service, email string, start time.Time, end time.Time) (*Calendar, error) {
	events, err := api.Events.List(email).
		AlwaysIncludeEmail(false).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(start.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		MaxResults(100).
		TimeZone("UTC").
		Q("out").
		Do()

	if err != nil {
		return nil, err
	}

	if len(events.Items) == 0 {
		return &Calendar{}, nil
	}

	out := &Calendar{}

	for _, item := range events.Items {
		start := item.Start.DateTime
		if start == "" {
			start = item.Start.Date + "T00:00:00Z"
		}

		end := item.End.DateTime
		if end == "" {
			end = item.End.Date + "T00:00:00Z"
		}

		startTime, err := time.Parse(time.RFC3339, start)
		if err != nil {
			return nil, err
		}

		endTime, err := time.Parse(time.RFC3339, end)
		if err != nil {
			return nil, err
		}

		out.Items = append(out.Items, &CalendarItem{Start: startTime, End: endTime})
	}

	return out, nil
}

func (c *CalendarAPI) getAPI(credsFile, tokFile string) (*calendar.Service, error) {
	b, err := ioutil.ReadFile(credsFile)
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, err
	}
	client := getClient(tokFile, config)

	return calendar.New(client)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(tokFile string, config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
