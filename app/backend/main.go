package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"appengine"
)

func init() {
	http.HandleFunc("/query", query)
}

func query(w http.ResponseWriter, r *http.Request) {
	lng, err := strconv.ParseFloat(r.FormValue("lng"), 64)
	if err != nil {
		http.Error(w, "Longitude could not be parsed", http.StatusBadRequest)
		return
	}
	lat, err := strconv.ParseFloat(r.FormValue("lat"), 64)
	if err != nil {
		http.Error(w, "Latitude could not be parsed", http.StatusBadRequest)
		return
	}
	c := appengine.NewContext(r)
	stations, err := getHubwayData(c)
	if err != nil {
		log.Panicf("error getting Hubway station data: %v", err)
		return
	}
	// TODO refactor function
	stations.good()
	point := GeoPoint{Lat: lat, Lng: lng}
	nearestStations := stations.closestStationsTo(&point, 5)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(nearestStations)
	return
}
