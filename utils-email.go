package main

import (
	"fmt"
	"os"

	// https://dev.mailjet.com/email/guides/send-api-v31
	mailjet "github.com/mailjet/mailjet-apiv3-go/v3"
)

var mailjetClient *mailjet.Client

func sendEmail(name, email, subject, plaintext, html string) error {

	if mailjetClient == nil {
		m, err := newMailjetClient()
		if err != nil {
			return fmt.Errorf("mailjet init: %w", err)
		}
		mailjetClient = m
	}

	var localDisclaimerPlain, localDisclaimerHTML string
	if isLocal() {
		localDisclaimerPlain = "(Running on localhost)\n\n"
		localDisclaimerHTML = "<p>(Running on localhost)</p>\n<br/>\n"
	}

	messages := mailjet.MessagesV31{
		Info: []mailjet.InfoMessagesV31{
			{
				From: &mailjet.RecipientV31{
					Email: "mailjet@crowdspec.dev", // TODO move to config
					Name:  "CrowdSpec Server",
				},
				To: &mailjet.RecipientsV31{
					mailjet.RecipientV31{
						Email: email,
						Name:  name,
					},
				},
				Subject:  subject,
				TextPart: localDisclaimerPlain + plaintext,
				HTMLPart: localDisclaimerHTML + html,
			},
		},
	}

	response, err := mailjetClient.SendMailV31(&messages)

	if response == nil || len(response.ResultsV31) == 0 ||
		response.ResultsV31[0].Status != "success" {
		return fmt.Errorf("email not sent: %w", err)
	}

	return err

}

func newMailjetClient() (*mailjet.Client, error) {
	var err error

	if mailjetSecretKey == "" {
		if isAppEngine() {
			mailjetSecretKey, err = loadSecret(os.Getenv("MAILJET_SECRET"))
			if err != nil {
				return nil, fmt.Errorf("loading MAILJET secret key: %w", err)
			}
		} else {
			return nil, fmt.Errorf("MAILJET secret key undefined")
		}
	}

	return mailjet.NewMailjetClient(mailjetAPIKey, mailjetSecretKey), nil
}
