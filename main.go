package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/genuinetools/1up/version"
	"github.com/genuinetools/pkg/cli"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	gmail "google.golang.org/api/gmail/v1"
)

const (
	tokenFile = "/tmp/token.json"
	gmailUser = "me"

	inboxLabel      = "inbox"
	goodLabel       = "1up/good"
	badLabel        = "1up/bad"
	quarantineLabel = "1up/quarantine"
)

var (
	credsFile string

	interval time.Duration
	once     bool

	debug bool
)

func main() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "1up"
	p.Description = "A custom Gmail spam filter bot"

	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)
	p.FlagSet.StringVar(&credsFile, "creds-file", os.Getenv("GMAIL_CREDENTIAL_FILE"), "Gmail credential file (or env var GMAIL_CREDENTIAL_FILE)")
	p.FlagSet.StringVar(&credsFile, "f", os.Getenv("GMAIL_CREDENTIAL_FILE"), "Gmail credential file (or env var GMAIL_CREDENTIAL_FILE)")

	p.FlagSet.DurationVar(&interval, "interval", 5*time.Minute, "update interval (ex. 5ms, 10s, 1m, 3h)")
	p.FlagSet.DurationVar(&interval, "i", 5*time.Minute, "update interval (ex. 5ms, 10s, 1m, 3h)")
	p.FlagSet.BoolVar(&once, "once", false, "run once and exit, do not run as a daemon")

	p.FlagSet.BoolVar(&debug, "debug", false, "enable debug logging")
	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if credsFile == "" {
			return errors.New("Gmail credential file cannot be empty")
		}

		// Make sure the file exists.
		if _, err := os.Stat(credsFile); os.IsNotExist(err) {
			return fmt.Errorf("Credential file %s does not exist", credsFile)
		}

		return nil
	}

	// Set the main program action.
	p.Action = func(ctx context.Context, repos []string) error {
		ticker := time.NewTicker(interval)

		// On ^C, or SIGTERM handle exit.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		go func() {
			for sig := range c {
				logrus.Infof("Received %s, exiting.", sig.String())
				ticker.Stop()
				cancel()
				os.Exit(0)
			}
		}()

		// Read the credentials file.
		b, err := ioutil.ReadFile(credsFile)
		if err != nil {
			logrus.Fatalf("Read client secret file %s failed: %v", credsFile, err)
		}

		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(b,
			// Read, send, delete, and manage your email.
			gmail.MailGoogleComScope)
		if err != nil {
			logrus.Fatalf("Parsing client secret file to config failed: %v", err)
		}

		// Get the client from the config.
		client, err := getClient(ctx, config)
		if err != nil {
			logrus.Fatalf("Creating client failed: %v", err)
		}

		// Create the service for the Gmail client.
		api, err := gmail.New(client)
		if err != nil {
			logrus.Fatalf("Creating Gmail client failed: %v", err)
		}

		// Get the labels for the user to make sure we have one for the bad label
		// and to find the correct label ID for "inbox".
		l, err := api.Users.Labels.List(gmailUser).Do()
		if err != nil {
			logrus.Fatalf("Listing labels failed: %v", err)
		}

		var goodLabelID, badLabelID, quarantineLabelID, inboxLabelID string

		logrus.Infof("Processing %d labels...", len(l.Labels))
		for _, label := range l.Labels {
			switch strings.ToLower(label.Name) {
			case goodLabel:
				goodLabelID = label.Id
			case badLabel:
				badLabelID = label.Id
			case quarantineLabel:
				quarantineLabelID = label.Id
			case inboxLabel:
				inboxLabelID = label.Id
			}

			if len(goodLabelID) > 0 && len(badLabelID) > 0 && len(quarantineLabelID) > 0 && len(inboxLabelID) > 0 {
				// We have everything we need so we can exit the loop.
				break
			}
		}

		// If the good label does not exist, create it.
		if len(goodLabelID) <= 0 {
			label, err := api.Users.Labels.Create(gmailUser, &gmail.Label{Name: goodLabel}).Do()
			if err != nil {
				logrus.Fatalf("Creating label %s failed: %v", goodLabel, err)
			}
			logrus.Infof("Created label %s", goodLabel)
			goodLabelID = label.Id
		}

		// If the bad label does not exist, create it.
		if len(badLabelID) <= 0 {
			label, err := api.Users.Labels.Create(gmailUser, &gmail.Label{Name: badLabel}).Do()
			if err != nil {
				logrus.Fatalf("Creating label %s failed: %v", badLabel, err)
			}
			logrus.Infof("Created label %s", badLabel)
			badLabelID = label.Id
		}

		// If the quarantine label does not exist, create it.
		if len(quarantineLabelID) <= 0 {
			label, err := api.Users.Labels.Create(gmailUser, &gmail.Label{Name: quarantineLabel}).Do()
			if err != nil {
				logrus.Fatalf("Creating label %s failed: %v", quarantineLabel, err)
			}
			logrus.Infof("Created label %s", quarantineLabel)
			quarantineLabelID = label.Id
		}

		run := func() {
			// Get the messages in the good and bad email labels.
			goodEmails, err := getMessagesForLabel(api, goodLabelID, goodLabel)
			if err != nil {
				logrus.Fatal(err)
			}
			badEmails, err := getMessagesForLabel(api, badLabelID, badLabel)
			if err != nil {
				logrus.Fatal(err)
			}

			// Train the classifier.
			classifier, err := trainClassifier(goodEmails, badEmails)
			if err != nil {
				logrus.Fatalf("Training classifier failed: %v", err)
			}

			// Get the messages in the inbox.
			r, err := api.Users.Messages.List(gmailUser).LabelIds(inboxLabelID).MaxResults(200).Do()
			if err != nil {
				logrus.Fatalf("Listing messages failed: %v", err)
			}

			logrus.Infof("Processing %d messages in %s...", len(r.Messages), inboxLabel)
			for _, m := range r.Messages {
				// Get the message.
				msg, err := api.Users.Messages.Get(gmailUser, m.Id).Format("full").Do()
				if err != nil {
					logrus.Fatalf("Getting message %s failed: %v", m.Id, err)
				}

				logrus.Debugf("snippet %s: %s", msg.Id, msg.Snippet)

				if msg.Payload == nil {
					// Continue if the message has no payload.
					continue
				}

				body := getMessageBody(msg)

				logrus.Debugf("message: %s", body)

				// Classify the message.
				scores, probs, isBad, err := classifier.classifyMessage(body)
				if err != nil {
					logrus.Warnf("Classifying message %s failed: %v", msg.Id, err)
				}

				// If the message is bad label it as quarantine.
				if isBad {
					logrus.Infof("Labeling bad message %s: scores: %#v -> probs: %#v -> snippet: %s", msg.Id, scores, probs, msg.Snippet)

					if _, err := api.Users.Messages.Modify(gmailUser, msg.Id,
						&gmail.ModifyMessageRequest{
							// Add the quarantine label.
							AddLabelIds: []string{quarantineLabelID},
							// Remove the inbox label.
							RemoveLabelIds: []string{inboxLabelID},
						}).Do(); err != nil {
						logrus.Warnf("Adding labels to message %s failed: %v", msg.Id, err)
					}

					logrus.Infof("Labeled message %s as %s", msg.Id, quarantineLabel)
				}
			}
		}

		// If the user passed the once flag, just do the run once and exit.
		if once {
			run()
			return nil
		}

		logrus.Infof("Starting bot to run every %s...", interval)
		for range ticker.C {
			run()
		}

		return nil
	}

	// Run our program.
	p.Run()
}
