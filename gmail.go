package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	gmail "google.golang.org/api/gmail/v1"
)

// getClient retrieves a token, saves the token, then returns the generated client.
func getClient(ctx context.Context, config *oauth2.Config) (*http.Client, error) {
	// Try reading the token from the file.
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		logrus.Warnf("Getting token from file failed: %v", err)

		// Could not get the token from the file, try reading it from the web.
		tok, err = getTokenFromWeb(ctx, config)
		if err != nil {
			return nil, err
		}

		// Save the token from the web.
		if err := saveToken(tokenFile, tok); err != nil {
			return nil, err
		}
	}

	return config.Client(ctx, tok), nil
}

// getTokenFromWeb requests a token from the web, then returns the retrieved token.
func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token from web: %v", err)
	}

	return tok, nil
}

// tokenFromFile retrieves a token from a local file.
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

// saveToken saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	logrus.Infof("Saving credential file to: %s", path)

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

func getMessageBody(msg *gmail.Message) (body string) {
	if msg.Payload.Body != nil && len(msg.Payload.Body.Data) > 0 {
		b, err := base64.StdEncoding.DecodeString(msg.Payload.Body.Data)
		if err != nil {
			logrus.Debugf("Parsing base64 encoded body for message ID %s failed: %v", msg.Id, err)
		} else {
			body += " " + string(b)
		}
	}

	// TODO(jessfraz): actually parse the parts smartly by
	// figuring out which are important and which are signatures, etc.
	for k, part := range msg.Payload.Parts {
		if part.Body != nil && len(part.Body.Data) > 0 {
			b, err := base64.StdEncoding.DecodeString(part.Body.Data)
			if err != nil {
				logrus.Debugf("Parsing base64 encoded body for message ID %s part %d failed: %v", msg.Id, k, err)
				continue
			}

			body += " " + string(b)
		}
	}

	return body
}

func getMessagesForLabel(api *gmail.Service, labelID, label string) ([]string, error) {
	r, err := api.Users.Messages.List(gmailUser).LabelIds(labelID).MaxResults(500).Do()
	if err != nil {
		return nil, fmt.Errorf("Listing messages for label %s failed: %v", label, err)
	}

	messages := []string{}

	logrus.Infof("Processing %d messages in %s...", len(r.Messages), label)
	for _, m := range r.Messages {
		// Get the message.
		msg, err := api.Users.Messages.Get(gmailUser, m.Id).Format("full").Do()
		if err != nil {
			return nil, fmt.Errorf("Getting message %s in label %s failed: %v", m.Id, label, err)
		}

		logrus.Debugf("snippet %s: %s", msg.Id, msg.Snippet)

		if msg.Payload == nil {
			// Continue if the message has no payload.
			continue
		}

		messages = append(messages, getMessageBody(msg))
	}

	return messages, nil
}
