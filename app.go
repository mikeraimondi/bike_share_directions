package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
)

var googleKey string

func init() {
	googleKey = os.Getenv("GKEY")
}

func main() {
	http.Handle("/bower_components/", http.StripPrefix("/bower_components/", http.FileServer(http.Dir("bower_components"))))
	http.Handle("/", http.FileServer(http.Dir("frontend")))

	http.HandleFunc("/query", root)
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func root(w http.ResponseWriter, r *http.Request) {
	lng, err := strconv.ParseFloat(r.FormValue("lng"), 64)
	if err != nil {
		http.Error(w, "Longitude could not be parsed", http.StatusBadRequest)
	}
	lat, err := strconv.ParseFloat(r.FormValue("lat"), 64)
	if err != nil {
		http.Error(w, "Latitude could not be parsed", http.StatusBadRequest)
	}
	point := GeoPoint{Lat: lat, Lng: lng}
	// TODO cache
	stations, err := getHubwayData(&http.Transport{})
	if err != nil {
		log.Panicf("error getting Hubway station data: %v", err)
		return
	}
	// TODO refactor function
	stations.good()
	nearestStations := stations.closestStationsTo(&point, 10)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(nearestStations)
	return
}
