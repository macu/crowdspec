#!/bin/sh

# Prompt for username and password
read -p "Enter a test username [test]: " username
read -sp "Enter a test password [test]: " password

# Output newline following password
echo

# Default each to "test"
username=${username:-test}
password=${password:-test}

# Record build date
sh update-build-date.sh

# Build frontend
npm run prod || { echo 'Client code compilation failed.' ; exit 1; }

# Run backend and init db and user
go run *.go -initDB -createUser -username="$username" -password="$password" -email=test@test.com
