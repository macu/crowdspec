#!/bin/sh

sh update-build-date.sh
npm run prod
go run *.go
