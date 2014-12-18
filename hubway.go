package app

import (
	"io/ioutil"
	"math"
	"net/url"
	"strconv"
	"sync"
	"time"

	"appengine"
	"appengine/memcache"
	"appengine/urlfetch"
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

func (sl *StationList) closestStationTo(point *appengine.GeoPoint) *Station {
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

var cacheLock sync.Mutex

func getHubwayData(c *appengine.Context) (hwData []byte, err error) {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	if item, err := memcache.Get(*c, "hubway"); err == memcache.ErrCacheMiss {
		(*c).Infof("Hubway station info cache miss")
		u, _ := url.Parse("www.thehubway.com/data/stations/bikeStations.xml")
		u.Scheme = "https"
		client := urlfetch.Client(*c)
		resp, err := client.Get(u.String())
		if err != nil {
			return hwData, err
		}
		if hwData, err = ioutil.ReadAll(resp.Body); err != nil {
			return hwData, err
		}
		newItem := &memcache.Item{
			Key:        "hubway",
			Value:      hwData,
			Expiration: time.Minute,
		}
		if err := memcache.Set(*c, newItem); err != nil {
			return hwData, err
		}
		return hwData, nil
	} else if err != nil {
		return hwData, err
	} else {
		hwData = item.Value
		return hwData, nil
	}
}
