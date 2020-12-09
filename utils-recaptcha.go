package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

func verifyRecaptcha(r *http.Request) (bool, error) {

	var err error

	if recaptchaSecretKey == "" {
		if isAppEngine() {
			recaptchaSecretKey, err = loadSecret(os.Getenv("RECAPTCHA_SECRET"))
			if err != nil {
				return false, fmt.Errorf("loading reCAPTCHA secret key: %w", err)
			}
		} else {
			return false, fmt.Errorf("reCAPTCHA secret key undefined")
		}
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	form := url.Values{}
	form.Set("secret", recaptchaSecretKey)
	form.Set("response", r.FormValue("g-recaptcha-response"))
	form.Set("remoteip", ip)

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	res, err := client.PostForm("https://www.google.com/recaptcha/api/siteverify", form)
	if err != nil {
		return false, fmt.Errorf("error fetching URL: %w", err)
	}

	defer res.Body.Close()

	var response = struct {
		Success bool `json:"success"`
	}{}

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return false, fmt.Errorf("error decoding response: %w", err)
	}

	return response.Success, nil
}