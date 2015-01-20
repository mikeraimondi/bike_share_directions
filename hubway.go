package main

import (
	"encoding/xml"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

// A StationList is a collection of Hubway stations
type StationList struct {
	Stations []Station `xml:"station"`
}

// A Station is a single Hubway station
type Station struct {
	ID                 uint64  `xml:"id"`
	Name               string  `xml:"name"`
	TerminalName       string  `xml:"terminalName"`
	LastCommWithServer uint64  `xml:"lastCommWithServer"`
	Lat                float64 `xml:"lat"`
	Lng                float64 `xml:"long"`
	Installed          bool    `xml:"installed"`
	Locked             bool    `xml:"locked"`
	InstallDate        uint64  `xml:"installDate"`
	RemovalDate        string  `xml:"removalDate"`
	Temporary          bool    `xml:"temporary"`
	Public             bool    `xml:"public"`
	Bikes              uint16  `xml:"nbBikes"`
	EmptyDocks         uint16  `xml:"nbEmptyDocks"`
	LatestUpdateTime   uint64  `xml:"latestUpdateTime"`
}

func (sl *StationList) good() {
	a := sl.Stations[:0]
	for _, station := range sl.Stations {
		if station.Bikes > 0 && station.Installed && !station.Locked && station.Public && len(station.RemovalDate) == 0 {
			a = append(a, station)
		}
	}
	sl.Stations = a
}

func (sl *StationList) closestStationTo(point *GeoPoint) *Station {
	best := sl.Stations[0]
	for _, station := range sl.Stations {
		if math.Abs(point.Lat-station.Lat)+math.Abs(point.Lng-station.Lng) < math.Abs(point.Lat-best.Lat)+math.Abs(point.Lng-best.Lng) {
			best = station
		}
	}
	return &best
}

func (s *Station) stringCoords() string {
	return strconv.FormatFloat(s.Lat, 'f', -1, 64) + "," + strconv.FormatFloat(s.Lng, 'f', -1, 64)
}

// TODO extract to library
func getHubwayData(transport http.RoundTripper) (stations *StationList, err error) {
	u, _ := url.Parse("www.thehubway.com/data/stations/bikeStations.xml")
	u.Scheme = "https"
	client := http.Client{Transport: transport}
	resp, err := client.Get(u.String())
	if err != nil {
		return stations, err
	}
	if err := xml.NewDecoder(resp.Body).Decode(&stations); err != nil {
		return stations, err
	}
	return stations, nil
}
