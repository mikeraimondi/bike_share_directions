package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

var securePort string
var insecurePort string

func init() {
	insecurePort = os.Getenv("PORT")
	securePort = os.Getenv("SSLPORT")
}

func main() {
	http.Handle("/bower_components/", http.StripPrefix("/bower_components/", http.FileServer(http.Dir("bower_components"))))
	http.Handle("/", http.FileServer(http.Dir("frontend")))

	http.HandleFunc("/query", root)
	// go func() {
	// 	if err := http.ListenAndServeTLS(":"+securePort, "cert.pem", "key.pem", nil); err != nil {
	// 		log.Fatal("ListenAndServeTLS: ", err)
	// 	}
	// }()
	// if err := http.ListenAndServe(":"+insecurePort, http.HandlerFunc(secureRedirect)); err != nil {
	// 	log.Fatal("ListenAndServe: ", err)
	// }
	if err := http.ListenAndServe(":"+insecurePort, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func root(w http.ResponseWriter, r *http.Request) {
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
	point := GeoPoint{Lat: lat, Lng: lng}
	// TODO cache
	stations, err := getHubwayData(&http.Transport{})
	if err != nil {
		log.Panicf("error getting Hubway station data: %v", err)
		return
	}
	// TODO refactor function
	stations.good()
	nearestStations := stations.closestStationsTo(&point, 5)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(nearestStations)
	return
}

func secureRedirect(w http.ResponseWriter, r *http.Request) {
	host, _, _ := net.SplitHostPort(r.Host)
	http.Redirect(w, r, "https://"+host+":443"+r.RequestURI, http.StatusMovedPermanently)
}
