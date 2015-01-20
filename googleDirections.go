package main

type Direction struct {
	Status string  `json:"status"`
	Routes []Route `json:"routes"`
}

type Route struct {
	Bounds Bound `json:"bounds"`
	Legs   []Leg `json:"legs"`
}

type Bound struct {
	Northeast GeoPoint `json:"northeast"`
	Southwest GeoPoint `json:"southwest"`
}

type Leg struct {
	Distance      Distance `json:"distance"`
	Duration      Duration `json:"duration"`
	EndAddress    string   `json:"end_address"`
	EndLocation   GeoPoint `json:"end_location"`
	StartAddress  string   `json:"start_address"`
	StartLocation GeoPoint `json:"start_location"`
	Steps         []Step   `json:"steps"`
}

type Distance struct {
	Text  string `json:"text"`
	Value int64  `json:"value"`
}

type Duration struct {
	Text  string `json:"text"`
	Value int64  `json:"value"`
}

type Step struct {
	Distance         Distance `json:"distance"`
	Duration         Duration `json:"duration"`
	EndLocation      GeoPoint `json:"end_location"`
	HTMLInstructions string   `json:"html_instructions"`
	Polyline         Polyline `json:"polyline"`
	StartLocation    GeoPoint `json:"start_location"`
	TravelMode       string   `json:"travel_mode"`
	Maneuver         string   `json:"maneuver"`
	Steps            []Step   `json:"steps"`
}

type Polyline struct {
	Points string
}

type GeoPoint struct {
	Lat float64
	Lng float64
}
