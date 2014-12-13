package app

import "appengine"

type Geocode struct {
	Status  string          `json:"status"`
	Results []GeocodeResult `json:"results"`
	context *appengine.Context
	address string
}

type GeocodeResult struct {
	FormattedAddress string          `json:"formatted_address"`
	Geometry         GeocodeGeometry `json:"geometry"`
}

type GeocodeGeometry struct {
	Bounds   Bound              `json:"bounds"`
	Location appengine.GeoPoint `json:"location"`
}
