package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
)

var client *http.Client

func init() {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(pemCerts)
	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: pool},
		},
	}
}

func main() {
	log.Println("App server started")
	http.Handle("/", http.FileServer(http.Dir("dist")))

	http.HandleFunc("/query", root)
	// go func() {
	// 	if err := http.ListenAndServeTLS(":"+securePort, "cert.pem", "key.pem", nil); err != nil {
	// 		log.Fatal("ListenAndServeTLS: ", err)
	// 	}
	// }()
	// if err := http.ListenAndServe(":"+insecurePort, http.HandlerFunc(secureRedirect)); err != nil {
	// 	log.Fatal("ListenAndServe: ", err)
	// }
	if err := http.ListenAndServe(":80", nil); err != nil {
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