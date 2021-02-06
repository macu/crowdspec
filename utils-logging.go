package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/logging"
)

var (
	appEngineProjectName   = os.Getenv("GOOGLE_CLOUD_PROJECT")
	appEngineLoggingClient *logging.Client
	appEngineLogger        *logging.Logger
	appEngineErrorClient   *errorreporting.Client
)

func initAppEngineLoggingClient() {
	if appEngineLoggingClient == nil {
		ctx := context.Background()
		var err error
		// Logs appear in Cloud Console in Logging
		appEngineLoggingClient, err = logging.NewClient(ctx, appEngineProjectName)
		if err != nil {
			log.Println(err)
			appEngineLoggingClient = nil
		}
	}
}

func initAppEngineLogger() {
	initAppEngineLoggingClient()
	if appEngineLoggingClient != nil {
		if appEngineLogger == nil {
			appEngineLogger = appEngineLoggingClient.Logger("logger")
		}
	}
}

func initAppEngineErrorClient() {
	if appEngineErrorClient == nil {
		ctx := context.Background()
		var err error
		// Errors appear in Cloud Console both in Error Reporting and in Logging
		appEngineErrorClient, err = errorreporting.NewClient(ctx, appEngineProjectName, errorreporting.Config{})
		if err != nil {
			log.Println(err)
			appEngineErrorClient = nil
		}
	}
}

func logDefault(r *http.Request, jsonPayload interface{}) {
	if isAppEngine() {
		initAppEngineLogger()
		if appEngineLogger != nil {
			appEngineLogger.Log(logging.Entry{
				Severity: logging.Default,
				Payload:  jsonPayload,
				HTTPRequest: &logging.HTTPRequest{
					Request: r,
				},
			})
		}
	}

	_, fn, line, _ := runtime.Caller(1)
	log.Printf("[default] %s:%d %v", fn, line, jsonPayload)
}

func logNotice(r *http.Request, jsonPayload interface{}) {
	if isAppEngine() {
		initAppEngineLogger()
		if appEngineLogger != nil {
			appEngineLogger.Log(logging.Entry{
				Severity: logging.Notice,
				Payload:  jsonPayload,
				HTTPRequest: &logging.HTTPRequest{
					Request: r,
				},
			})
		}
	}

	_, fn, line, _ := runtime.Caller(1)
	log.Printf("[notice] %s:%d %v", fn, line, jsonPayload)
}

func logError(r *http.Request, userID uint, err error) {
	if err == nil {
		return
	}

	if isAppEngine() {
		initAppEngineErrorClient()
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
	// Continue even if err is nil

	if isAppEngine() {
		initAppEngineErrorClient()
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
