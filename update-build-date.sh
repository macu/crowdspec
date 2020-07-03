#!/bin/sh

DATETIME=$(date +'%Y%m%d%H%M')

# Replace versionStamp in local dev file
sed -i -E "/\"versionStamp\":/ s/\"versionStamp\": \"[0-9]*\"/\"versionStamp\": \"$DATETIME\"/" env.json
rm env.json-E

# Replace VERSION_STAMP in app.yaml
sed -i -E "/VERSION_STAMP:/ s/VERSION_STAMP: \"[0-9]*\"/VERSION_STAMP: \"$DATETIME\"/" app.yaml
rm app.yaml-E
