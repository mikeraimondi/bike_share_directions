package app

import (
	"net/http"

	"appengine"
)

type trip struct {
	origin      endpoint
	destination endpoint
}

type endpoint struct {
	address             string
	geocode             Geocode
	nearestStation      Station
	directionsToStation Direction
	origin              bool
	context             *appengine.Context
	client              *http.Client
}
