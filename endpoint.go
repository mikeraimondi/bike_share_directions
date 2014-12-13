package app

import "appengine"

type endpoint struct {
	address             string
	geocode             Geocode
	nearestStation      Station
	directionsToStation Direction
	context             *appengine.Context
}
