package main

type Geocode struct {
	Status  string          `json:"status"`
	Results []GeocodeResult `json:"results"`
}

type GeocodeResult struct {
	FormattedAddress string          `json:"formatted_address"`
	Geometry         GeocodeGeometry `json:"geometry"`
}

type GeocodeGeometry struct {
	Bounds   Bound    `json:"bounds"`
	Location GeoPoint `json:"location"`
}
