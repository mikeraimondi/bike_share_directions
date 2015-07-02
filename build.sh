#!/bin/sh

CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -installsuffix netgo .
node_modules/.bin/bower install --config.interactive=false
node_modules/.bin/gulp
docker build -t mikeraimondi/bike_share_directions:latest .
