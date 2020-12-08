package main

import (
	"fmt"
	"os"

	mailjet "github.com/mailjet/mailjet-apiv3-go/v3"
)

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
