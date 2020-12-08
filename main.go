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
	"regexp"
	"strings"

	"github.com/gorilla/mux" // (BSD-3-Clause) https://github.com/gorilla/mux/blob/master/LICENSE

	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib"
)

// Represents local env.json config
type config struct {
	DBUser           string `json:"dbUser"`
	DBPass           string `json:"dbPass"`
	DBName           string `json:"dbName"`
	HTTPPort         string `json:"httpPort"`
	ReSiteKey        string `json:"recaptchaSiteKey"`
	ReSecretKey      string `json:"recaptchaSecretKey"`
	MailjetAPIKey    string `json:"mailjetApiKey"`
	MailjetSecretKey string `json:"mailjetSecretKey"`
	VersionStamp     string `json:"versionStamp"`
}

// reCAPTCHA site keys
var recaptchaSiteKey, recaptchaSecretKey string

// Mailjet API keys
var mailjetAPIKey, mailjetSecretKey string

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

	initDB := flag.Bool("initDB", false, "Initialize a fresh database")
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

		// Port number comes from env on App Engine
		config.HTTPPort = os.Getenv("PORT")

		// Site key comes from env
		recaptchaSiteKey = os.Getenv("RECAPTCHA_SITE_KEY")
		mailjetAPIKey = os.Getenv("MAILJET_API_KEY")

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

		// API keys and secret keys come from env.json
		recaptchaSiteKey = config.ReSiteKey
		recaptchaSecretKey = config.ReSecretKey
		mailjetAPIKey = config.MailjetAPIKey
		mailjetSecretKey = config.MailjetSecretKey

		// Version stamp comes from env.json
		cacheControlVersionStamp = config.VersionStamp

	}

	// Confirm connection
	err = db.Ping()
	if err != nil {
		logErrorFatal(err)
	}

	if isLocal() && *initDB {
		// Load initializing SQL
		initFileContents, err := ioutil.ReadFile("sql/init.pgsql")
		if err != nil {
			logErrorFatal(err)
		}

		// Remove comments
		blockCommentMatcher := regexp.MustCompile("(?s)/\\*.*?\\*/")
		// Match dashed comment lines and trailing dashed comments
		dashedCommentMatcher := regexp.MustCompile("(?m)(^\\s*--.*$[\r\n]*)|(\\s*--.*$)")
		// Remove block comments first
		withoutComments := dashedCommentMatcher.ReplaceAllString(
			blockCommentMatcher.ReplaceAllString(string(initFileContents), ""), "")

		// Split into statements
		lines := strings.Split(withoutComments, ";")

		// Execute each statement
		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			_, err = db.Exec(line)
			if err != nil {
				logErrorFatal(err)
			}
		}
		log.Println("Database initialized")
	}

	if isLocal() && *createNewUser {
		if _, err := createUser(db, *newUserUsername, *newUserPassword, *newUserEmail); err != nil {
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
	r.HandleFunc("/request-password-reset", makeRequestPasswordResetHandler(db))
	r.HandleFunc("/reset-password", makeResetPasswordHandler(db))
	r.HandleFunc("/logout", authenticate(logoutHandler))

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

	if appEngineErrorClient != nil {
		// Flush pending logs
		appEngineErrorClient.Close()
	}
}

var indexTemplate = template.Must(template.ParseFiles("html/index.html"))

func indexHandler(db *sql.DB, userID uint, w http.ResponseWriter, r *http.Request) {
	row := db.QueryRow("SELECT username FROM user_account WHERE id=$1", userID)
	var username string
	err := row.Scan(&username)
	if err != nil {
		logError(r, userID, fmt.Errorf("selecting username: %w", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	settings, err := loadUserSettings(db, userID)
	if err != nil {
		logError(r, userID, fmt.Errorf("loading user settings: %w", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	indexTemplate.Execute(w, struct {
		UserID       uint
		Username     string
		Settings     UserSettings
		VersionStamp string
		Local        bool
	}{userID, username, *settings, cacheControlVersionStamp, isLocal()})
}
