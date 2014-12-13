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

func gen(c *appengine.Context, addresses ...string) <-chan Geocode {
	out := make(chan Geocode)
	go func() {
		for _, address := range addresses {
			out <- Geocode{address: address, context: c}
		}
		close(out)
	}()
	return out
}

func geocode(in <-chan Geocode) <-chan Geocode {
	out := make(chan Geocode)
	go func() {
		for n := range in {
			u, _ := url.Parse("maps.googleapis.com/maps/api/geocode/json")
			u.Scheme = "https"
			q := u.Query()
			q.Set("key", googleKey)
			q.Set("address", n.address)
			u.RawQuery = q.Encode()
			client := urlfetch.Client(*n.context)
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
			out <- gc
		}
		close(out)
	}()
	return out
}

func merge(cs ...<-chan Geocode) <-chan Geocode {
	var wg sync.WaitGroup
	out := make(chan Geocode)

	output := func(c <-chan Geocode) {
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
	ch1 := geocode(in)
	ch2 := geocode(in)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var sl []Geocode
	for n := range merge(ch1, ch2) {
		sl = append(sl, n)
	}
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
