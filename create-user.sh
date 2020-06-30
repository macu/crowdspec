#!/bin/sh

# Prompt for username and password
read -p "Enter a username [test]: " username

read -p "Enter an email address [required]: " email

if [ "$email" = "" ]; then
	echo "Email required"
	exit 1
fi

read -sp "Enter a password [test]: " password

# Output newline following password
echo

# Default each to "test"
username=${username:-test}
password=${password:-test}

# Run backend and create user
go run *.go -createUser -username="$username" -password="$password" -email="$email" -exit
