#!/bin/sh

cd app/backend
go get
CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -installsuffix netgo -o bike_share_directions .
cd ../..
node_modules/.bin/bower install --config.interactive=false
node_modules/.bin/gulp
docker build -t mikeraimondi/bike_share_directions:latest .
rm app/backend/bike_share_directions
