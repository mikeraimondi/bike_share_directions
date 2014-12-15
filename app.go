package app

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"appengine"
	"appengine/urlfetch"
)

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/gd", googleDirections)
	http.HandleFunc("/hw", hubway)
}

func gen(c *appengine.Context, addresses ...string) <-chan endpoint {
	out := make(chan endpoint)
	go func() {
		for _, address := range addresses {
			out <- endpoint{address: address, context: c}
		}
		close(out)
	}()
	return out
}

func geocode(in <-chan endpoint) <-chan endpoint {
	out := make(chan endpoint)
	go func() {
		for endpoint := range in {
			u, _ := url.Parse("maps.googleapis.com/maps/api/geocode/json")
			u.Scheme = "https"
			q := u.Query()
			q.Set("key", googleKey)
			q.Set("address", endpoint.address)
			u.RawQuery = q.Encode()
			client := urlfetch.Client(*endpoint.context)
			resp, err := client.Get(u.String())
			if err != nil {
				// TODO error handling
				return
			}
			var gc Geocode
			if err := json.NewDecoder(resp.Body).Decode(&gc); err != nil {
				// TODO error handling
				return
			}
			endpoint.geocode = gc
			out <- endpoint
		}
		close(out)
	}()
	return out
}

func findNearestStation(in <-chan endpoint) <-chan endpoint {
	out := make(chan endpoint)
	go func() {
		for endpoint := range in {
			// TODO guard endpoint.geocode

			// TODO cache hubway results
			u, _ := url.Parse("www.thehubway.com/data/stations/bikeStations.xml")
			u.Scheme = "https"
			client := urlfetch.Client(*endpoint.context)
			resp, err := client.Get(u.String())
			if err != nil {
				// TODO error handling
				return
			}
			var sl StationList
			if err := xml.NewDecoder(resp.Body).Decode(&sl); err != nil {
				// TODO error handling
				return
			}
			// TODO refactor function
			sl.good()
			endpoint.nearestStation = *sl.closestStationTo(&endpoint.geocode.Results[0].Geometry.Location)
			out <- endpoint
		}
		close(out)
	}()
	return out
}

func stationDirections(in <-chan endpoint) <-chan endpoint {
	out := make(chan endpoint)
	go func() {
		for endpoint := range in {
			// TODO guard endpoint.geocode
			// TODO guard endpoint.nearestStation
			client := urlfetch.Client(*endpoint.context)
			u, _ := url.Parse("maps.googleapis.com/maps/api/directions/json")
			u.Scheme = "https"
			q := u.Query()
			q.Set("origin", endpoint.address)
			q.Set("destination", strconv.FormatFloat(endpoint.nearestStation.Lat, 'f', -1, 64)+","+strconv.FormatFloat(endpoint.nearestStation.Lng, 'f', -1, 64))
			q.Set("key", googleKey)
			q.Set("mode", "walking")
			// q.Set("departure_time", strconv.FormatInt(time.Now().Unix(), 10))
			u.RawQuery = q.Encode()
			resp, err := client.Get(u.String())
			if err != nil {
				// TODO error handling
				return
			}
			var d Direction
			if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
				// TODO error handling
				return
			}
			endpoint.directionsToStation = d
			out <- endpoint
		}
		close(out)
	}()
	return out
}

func merge(cs ...<-chan endpoint) <-chan endpoint {
	var wg sync.WaitGroup
	out := make(chan endpoint)

	output := func(c <-chan endpoint) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	in := gen(&c, r.FormValue("origin"), r.FormValue("destination"))
	ch1 := stationDirections(findNearestStation(geocode(in)))
	ch2 := stationDirections(findNearestStation(geocode(in)))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var s []Direction
	for endpoint := range merge(ch1, ch2) {
		s = append(s, endpoint.directionsToStation)
	}
	json.NewEncoder(w).Encode(s)
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
