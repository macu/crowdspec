package main

import (
	"bytes"
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

func logError(r *http.Request, userID *uint, err error) {
	if err == nil {
		return
	}

	if isAppEngine() {
		initAppEngineErrorClient()
		if appEngineErrorClient != nil {
			var user string
			if userID != nil {
				user = UintToA(*userID)
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

// ResponseTracker tracks bytes written in an HTTP response.
// Thanks https://www.reddit.com/r/golang/comments/ffwkh9/api_measure_response_size_in_bytes/fk6we0b
type ResponseTracker struct {
	http.ResponseWriter
	total int
}

// NewResponseTracker wraps the given writer in ResponseTracker.
func NewResponseTracker(w http.ResponseWriter) *ResponseTracker {
	return &ResponseTracker{ResponseWriter: w}
}

// WriteHeader counts the bytes written in headers.
// Always call this manually before calling Write
// to ensure the headers get counted.
func (t *ResponseTracker) WriteHeader(code int) {
	t.ResponseWriter.WriteHeader(code)
	var buf bytes.Buffer
	t.Header().Write(&buf)
	t.total += buf.Len()
}

// Write counts the bytes written.
func (t *ResponseTracker) Write(b []byte) (int, error) {
	n, err := t.ResponseWriter.Write(b)
	t.total += n
	return n, err
}
