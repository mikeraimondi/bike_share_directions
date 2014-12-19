package app

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sync"

	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/memcache"

	"appengine"
	"appengine/urlfetch"
)

func init() {
	http.HandleFunc("/", root)
}

func gen(c *appengine.Context, origin string, destination string) <-chan endpoint {
	out := make(chan endpoint)
	client := urlfetch.Client(*c)
	go func() {
		out <- endpoint{address: origin, origin: true, context: c, client: client}
		out <- endpoint{address: destination, origin: false, context: c, client: client}
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
			resp, err := endpoint.client.Get(u.String())
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

			// TODO cache expiration
			transport := httpcache.NewTransport(memcache.New(*endpoint.context))
			transport.Transport = endpoint.client.Transport
			stations, err := getHubwayData(transport)
			if err != nil {
				(*endpoint.context).Errorf("error getting Hubway station data: %v", err)
				// TODO cancel
				return
			}
			// TODO refactor function
			stations.good()
			endpoint.nearestStation = *stations.closestStationTo(&endpoint.geocode.Results[0].Geometry.Location)
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
			u, _ := url.Parse("maps.googleapis.com/maps/api/directions/json")
			u.Scheme = "https"
			q := u.Query()
			if endpoint.origin {
				q.Set("origin", endpoint.address)
				q.Set("destination", endpoint.nearestStation.stringCoords())
			} else {
				q.Set("origin", endpoint.nearestStation.stringCoords())
				q.Set("destination", endpoint.address)
			}
			q.Set("key", googleKey)
			q.Set("mode", "walking")
			// q.Set("departure_time", strconv.FormatInt(time.Now().Unix(), 10))
			u.RawQuery = q.Encode()
			resp, err := endpoint.client.Get(u.String())
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

func stationToStation(origin *endpoint, destination *endpoint) (d *Direction, err error) {
	u, _ := url.Parse("maps.googleapis.com/maps/api/directions/json")
	u.Scheme = "https"
	q := u.Query()
	q.Set("origin", origin.nearestStation.stringCoords())
	q.Set("destination", destination.nearestStation.stringCoords())
	q.Set("key", googleKey)
	q.Set("mode", "bicycling")
	u.RawQuery = q.Encode()
	resp, err := origin.client.Get(u.String())
	if err != nil {
		// TODO error handling
		return nil, err
	}
	// d Direction
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		// TODO error handling
		return nil, err
	}
	return d, nil
}

func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	in := gen(&c, r.FormValue("origin"), r.FormValue("destination"))
	ch1 := stationDirections(findNearestStation(geocode(in)))
	ch2 := stationDirections(findNearestStation(geocode(in)))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var t trip
	for endpoint := range merge(ch1, ch2) {
		if endpoint.origin {
			t.origin = endpoint
		} else {
			t.destination = endpoint
		}
	}
	d, _ := stationToStation(&t.origin, &t.destination)
	s := struct {
		DirectionsToStation   Direction
		InterStation          Direction
		DirectionsFromStation Direction
	}{
		t.origin.directionsToStation,
		*d,
		t.destination.directionsToStation,
	}
	json.NewEncoder(w).Encode(s)
	return
}
