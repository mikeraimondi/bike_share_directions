package app

import "appengine"

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
}
