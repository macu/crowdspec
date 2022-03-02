package main

import (
	"fmt"
	"os"
	"strings"

	// https://dev.mailjet.com/email/guides/send-api-v31
	mailjet "github.com/mailjet/mailjet-apiv3-go/v3"
)

const senderEmailAddress = "mailjet@crowdspec.dev"
const platformRecipientEmailAddress = "crowdspec.dev@gmail.com"

var mailjetClient *mailjet.Client

// plaintext email body is required
// html body may be blank
func sendEmail(name, email, subject, plaintext, html string, copyAdmin bool) error {

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

	var htmlPart string
	if strings.TrimSpace(html) != "" {
		htmlPart = localDisclaimerHTML + html
	}

	var bccRecipients = mailjet.RecipientsV31{}
	if copyAdmin {
		bccRecipients = mailjet.RecipientsV31{
			mailjet.RecipientV31{
				Email: platformRecipientEmailAddress,
				Name:  "CrowdSpec",
			},
		}
	}

	messages := mailjet.MessagesV31{
		Info: []mailjet.InfoMessagesV31{
			{
				From: &mailjet.RecipientV31{
					Email: senderEmailAddress,
					Name:  "CrowdSpec Server",
				},
				To: &mailjet.RecipientsV31{
					mailjet.RecipientV31{
						Email: email,
						Name:  name,
					},
				},
				Bcc:      &bccRecipients,
				Subject:  subject,
				TextPart: localDisclaimerPlain + plaintext,
				HTMLPart: htmlPart,
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

// plaintext email body is required
// html body may be blank
func sendEmailBcc(names []string, emails []string, subject, plaintext, html string) error {

	if len(names) != len(emails) {
		return fmt.Errorf("mismatch in length of names and emails")
	}

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

	var htmlPart string
	if strings.TrimSpace(html) != "" {
		htmlPart = localDisclaimerHTML + html
	}

	var bccRecipients = mailjet.RecipientsV31{}
	for i := 0; i < len(names); i++ {
		bccRecipients = append(bccRecipients, mailjet.RecipientV31{
			Email: emails[i],
			Name:  names[i],
		})
	}

	messages := mailjet.MessagesV31{
		Info: []mailjet.InfoMessagesV31{
			{
				From: &mailjet.RecipientV31{
					Email: senderEmailAddress,
					Name:  "CrowdSpec Server",
				},
				To: &mailjet.RecipientsV31{
					mailjet.RecipientV31{
						Email: platformRecipientEmailAddress,
						Name:  "CrowdSpec",
					},
				},
				Bcc:      &bccRecipients,
				Subject:  subject,
				TextPart: localDisclaimerPlain + plaintext,
				HTMLPart: htmlPart,
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
