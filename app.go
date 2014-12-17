package app

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"appengine"
	"appengine/memcache"
	"appengine/urlfetch"
)

func init() {
	http.HandleFunc("/", root)
}

func gen(c *appengine.Context, origin string, destination string) <-chan endpoint {
	out := make(chan endpoint)
	go func() {
		out <- endpoint{address: origin, origin: true, context: c}
		out <- endpoint{address: destination, origin: false, context: c}
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

			var hwData []byte
			// TODO some kind of lock?
			if item, err := memcache.Get(*endpoint.context, "hubway"); err == memcache.ErrCacheMiss {
				(*endpoint.context).Infof("item not in the cache")
				u, _ := url.Parse("www.thehubway.com/data/stations/bikeStations.xml")
				u.Scheme = "https"
				client := urlfetch.Client(*endpoint.context)
				resp, err := client.Get(u.String())
				if err != nil {
					(*endpoint.context).Errorf("error GET hubway XML: %v", err)
					return
				}
				// TODO error handling
				if hwData, err = ioutil.ReadAll(resp.Body); err != nil {
					(*endpoint.context).Errorf("error reading Hubway response: %v", err)
				}
				newItem := &memcache.Item{
					Key:   "hubway",
					Value: hwData,
				}
				if err := memcache.Set(*endpoint.context, newItem); err != nil {
					(*endpoint.context).Errorf("error setting item: %v", err)
				}
			} else if err != nil {
				(*endpoint.context).Errorf("error getting item: %v", err)
			} else {
				hwData = item.Value
			}
			var sl StationList
			if err := xml.Unmarshal(hwData, &sl); err != nil {
				(*endpoint.context).Errorf("error unmarshaling Hubway XML: %v", err)
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

func stationToStation(origin *endpoint, destination *endpoint) (d *Direction, err error) {
	client := urlfetch.Client(*origin.context)
	u, _ := url.Parse("maps.googleapis.com/maps/api/directions/json")
	u.Scheme = "https"
	q := u.Query()
	q.Set("origin", origin.nearestStation.stringCoords())
	q.Set("destination", destination.nearestStation.stringCoords())
	q.Set("key", googleKey)
	q.Set("mode", "bicycling")
	u.RawQuery = q.Encode()
	resp, err := client.Get(u.String())
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
