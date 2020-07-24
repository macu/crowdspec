package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"cloud.google.com/go/errorreporting"
)

var (
	appEngineProjectName = os.Getenv("GOOGLE_CLOUD_PROJECT")
	appEngineErrorClient *errorreporting.Client
)

func initErrorClient() {
	if appEngineErrorClient == nil {
		ctx := context.Background()
		var err error
		// Errors appear in Cloud Console both in Error Reporting and in Logging
		appEngineErrorClient, err = errorreporting.NewClient(ctx, appEngineProjectName, errorreporting.Config{})
		if err != nil {
			appEngineErrorClient = nil
		}
	}
}

func logError(r *http.Request, userID uint, err error) {
	if err == nil {
		return
	}

	if appengine {
		initErrorClient()
		if appEngineErrorClient != nil {
			var user string
			if userID > 0 {
				user = UintToA(userID)
			}
			appEngineErrorClient.Report(errorreporting.Entry{
				Error: err,
				Req:   r,
				User:  user,
			})
		}
	}

	_, fn, line, _ := runtime.Caller(1)
	log.Printf("[error] %s:%d %v", fn, line, err)
}

func logErrorFatal(err error) {
	if appengine {
		initErrorClient()
		if appEngineErrorClient != nil {
			appEngineErrorClient.Report(errorreporting.Entry{
				Error: err,
			})
			appEngineErrorClient.Close()
		}
	}

	_, fn, line, _ := runtime.Caller(1)
	log.Fatalln(fmt.Sprintf("[fatal] %s:%d %v", fn, line, err))
}
