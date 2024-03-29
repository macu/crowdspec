package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux" // (BSD-3-Clause) https://github.com/gorilla/mux/blob/master/LICENSE

	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib"
)

// Represents local env.json config
type config struct {
	DBUser            string `json:"dbUser"`
	DBPass            string `json:"dbPass"`
	DBName            string `json:"dbName"`
	HTTPPort          string `json:"httpPort"`
	AdminUserID       uint   `json:"adminUserId"`
	ReSiteKey         string `json:"recaptchaSiteKey"`
	ReSecretKey       string `json:"recaptchaSecretKey"`
	MailjetAPIKey     string `json:"mailjetApiKey"`
	MailjetSecretKey  string `json:"mailjetSecretKey"`
	YoutubeAPIKey     string `json:"youtubeApiKey"`
	HTTPClientReferer string `json:"httpClientReferer"`
	VersionStamp      string `json:"versionStamp"`
}

// reCAPTCHA site keys
var recaptchaSiteKey, recaptchaSecretKey string

// Mailjet API keys
var mailjetAPIKey, mailjetSecretKey string

// YouTube API keys
var youtubeAPIKey, httpClientReferer string

// Used to invalidate cache on compiled client resources
var cacheControlVersionStamp string

// Returns detection of appengine
func isAppEngine() bool {
	return os.Getenv("DEV_ENV") == "appengine"
}

// Returns non-detection of appengine
func isLocal() bool {
	return !isAppEngine()
}

func main() {

	if AtoBool(os.Getenv("MAINTENANCE_MODE")) {
		// Site down for databae upgrades
		maintenanceMode()
		os.Exit(0)
	}

	createNewUser := flag.Bool("createUser", false, "Whether to create a user on startup")
	newUserUsername := flag.String("username", "", "Login username for new user")
	newUserPassword := flag.String("password", "", "Password for new user")
	newUserEmail := flag.String("email", "", "Email address for new user")
	exitAfterExec := flag.Bool("exit", false, "Whether to exit after init or create user")
	flag.Parse()

	var config = &config{}
	var db *sql.DB
	var err error

	// Establish database connection
	if isAppEngine() {

		// Load DB password from Secret Manager
		secretName := os.Getenv("DB_PASS_SECRET")
		dbPass, err := loadSecret(secretName)
		if err != nil {
			logErrorFatal(err)
		}

		dataSource := fmt.Sprintf("host=%s user=%s password=%s database=%s",
			os.Getenv("DB_HOST"), os.Getenv("DB_USER"), dbPass, os.Getenv("DB_NAME"))

		db, err = sql.Open("pgx", dataSource)
		if err != nil {
			logErrorFatal(err)
		}

		adminUserID, err = AtoUint(os.Getenv("ADMIN_USER_ID"))
		if err != nil {
			logErrorFatal(fmt.Errorf("parsing admin ID: %w", err))
		}

		// Port number comes from env on App Engine
		config.HTTPPort = os.Getenv("PORT")

		// Site key comes from env
		recaptchaSiteKey = os.Getenv("RECAPTCHA_SITE_KEY")
		mailjetAPIKey = os.Getenv("MAILJET_API_KEY")
		youtubeAPIKey = os.Getenv("YOUTUBE_API_KEY")
		httpClientReferer = os.Getenv("HTTP_CLIENT_REFERER")

		// Version stamp comes from env
		cacheControlVersionStamp = os.Getenv("VERSION_STAMP")

	} else {
		// Running locally

		configContents, err := ioutil.ReadFile("env.json")
		if err != nil {
			logErrorFatal(err)
		}

		err = json.Unmarshal(configContents, config)
		if err != nil {
			logErrorFatal(err)
		}

		dataSource := fmt.Sprintf("host=%s user=%s password=%s database=%s",
			"localhost", config.DBUser, config.DBPass, config.DBName)

		db, err = sql.Open("pgx", dataSource)
		if err != nil {
			logErrorFatal(err)
		}

		adminUserID = config.AdminUserID

		// API keys and secret keys come from env.json
		recaptchaSiteKey = config.ReSiteKey
		recaptchaSecretKey = config.ReSecretKey
		mailjetAPIKey = config.MailjetAPIKey
		mailjetSecretKey = config.MailjetSecretKey
		youtubeAPIKey = config.YoutubeAPIKey
		httpClientReferer = config.HTTPClientReferer

		// Version stamp comes from env.json
		cacheControlVersionStamp = config.VersionStamp

	}

	// Confirm connection
	// err = db.Ping()
	// if err != nil {
	// 	logErrorFatal(err)
	// }

	if isLocal() && *createNewUser {
		if _, err := createUser(&http.Request{}, // request needed for r.Context()
			db, *newUserUsername, *newUserPassword, *newUserEmail); err != nil {
			logErrorFatal(err)
		}
		log.Println("New user created")
	}

	if *exitAfterExec {
		return
	}

	r := mux.NewRouter()

	if isLocal() {
		// Set up static resource routes
		// (These static directories are configured by app.yaml for App Engine)
		r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
		r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("img"))))
		r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))
	} else if isAppEngine() {
		// Set up AppEngine cron handlers
		r.HandleFunc("/cron/cleanup", makeCronHandler(db, cleanupHandler))
	}

	// set up authenticated routes
	authenticate := makeAuthenticator(db)

	r.HandleFunc("/login", makeLoginHandler(db))
	r.HandleFunc("/signup", makeRequestSignupHandler(db))
	r.HandleFunc("/activate-signup", makeActivateSignupHandler(db))
	r.HandleFunc("/request-password-reset", makeRequestPasswordResetHandler(db))
	r.HandleFunc("/reset-password", makeResetPasswordHandler(db))

	r.PathPrefix("/ajax/").HandlerFunc(authenticate(ajaxHandler))

	// All other paths go through index handler
	r.PathPrefix("/").HandlerFunc(authenticate(indexHandler))

	s := &http.Server{
		Addr:    ":" + config.HTTPPort,
		Handler: r,
	}

	log.Printf("Listening on port %s", config.HTTPPort)

	if err = s.ListenAndServe(); err != http.ErrServerClosed {
		logErrorFatal(err)
	}

	// Flush pending logs
	if appEngineLoggingClient != nil {
		appEngineLoggingClient.Close()
		appEngineLoggingClient = nil
	}
	if appEngineErrorClient != nil {
		appEngineErrorClient.Close()
		appEngineErrorClient = nil
	}
}

var indexTemplate = template.Must(template.ParseFiles("html/index.html"))

func indexHandler(db *sql.DB, userID *uint, w http.ResponseWriter, r *http.Request) {
	indexTemplate.Execute(w, struct {
		VersionStamp        string
		PasswordMinLength   uint
		SpecNameMaxLength   uint
		BlockTitleMaxLength uint
		URLMaxLength        uint
		Local               bool
	}{
		cacheControlVersionStamp,
		passwordMinLength,
		subspecNameMaxLen,
		blockTitleMaxLen,
		urlMaxLen,
		isLocal(),
	})
}

func maintenanceMode() {
	var port string
	if isAppEngine() {
		port = os.Getenv("PORT")
	} else {
		var config = &config{}
		configContents, err := ioutil.ReadFile("env.json")
		if err != nil {
			logErrorFatal(err)
		}
		err = json.Unmarshal(configContents, config)
		if err != nil {
			logErrorFatal(err)
		}
		port = config.HTTPPort
	}

	var maintenancePageTemplate *template.Template
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		if !isAjax(r) {
			if maintenancePageTemplate == nil {
				maintenancePageTemplate = template.Must(template.ParseFiles("html/maintenance.html"))
			}
			maintenancePageTemplate.Execute(w, nil)
		}
	})

	log.Printf("Listening on port %s", port)

	if err := http.ListenAndServe(":"+port, nil); err != http.ErrServerClosed {
		logErrorFatal(err)
	}

	// Flush pending logs
	if appEngineLoggingClient != nil {
		appEngineLoggingClient.Close()
		appEngineLoggingClient = nil
	}
	if appEngineErrorClient != nil {
		appEngineErrorClient.Close()
		appEngineErrorClient = nil
	}
}
