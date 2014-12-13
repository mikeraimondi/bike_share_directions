package app

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"appengine"
	"appengine/urlfetch"
)

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/gd", googleDirections)
	http.HandleFunc("/hw", hubway)
}

func geocode(address string, c *appengine.Context, gChan chan Geocode) {
	u, _ := url.Parse("maps.googleapis.com/maps/api/geocode/json")
	u.Scheme = "https"
	q := u.Query()
	q.Set("key", googleKey)
	q.Set("address", address)
	u.RawQuery = q.Encode()
	client := urlfetch.Client(*c)
	resp, err := client.Get(u.String())
	if err != nil {
		// TODO error handling
		return
	}
	gc := Geocode{}
	if err := json.NewDecoder(resp.Body).Decode(&gc); err != nil {
		// TODO error handling
		return
	}
	gChan <- gc
	return
}

func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	gChan := make(chan Geocode)

	go geocode(r.FormValue("origin"), &c, gChan)
	go geocode(r.FormValue("destination"), &c, gChan)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var sl []Geocode
	sl = append(sl, <-gChan)
	sl = append(sl, <-gChan)
	json.NewEncoder(w).Encode(sl)
	return
}

func googleDirections(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	u, err := url.Parse("maps.googleapis.com/maps/api/directions/json")
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	u.Scheme = "https"
	q := u.Query()
	q.Set("origin", r.FormValue("origin"))
	q.Set("destination", r.FormValue("destination"))
	q.Set("key", googleKey)
	q.Set("mode", "walking")
	q.Set("departure_time", strconv.FormatInt(time.Now().Unix(), 10))
	u.RawQuery = q.Encode()
	resp, err := client.Get(u.String())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var d Direction
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(d)
	return
}

func hubway(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	u, err := url.Parse("www.thehubway.com/data/stations/bikeStations.xml")
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	u.Scheme = "https"
	resp, err := client.Get(u.String())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var s StationList
	if err := xml.NewDecoder(resp.Body).Decode(&s); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.good()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(s)
	return
}
