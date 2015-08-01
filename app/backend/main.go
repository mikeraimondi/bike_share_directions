package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	client    *http.Client
	httpPort  = os.Getenv("HTTP_PORT")
	redisPool *redis.Pool
)

func init() {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(pemCerts)
	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: pool},
		},
	}

	redisPool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			redisServer := os.Getenv("REDIS_SERVER")
			if len(redisServer) == 0 {
				redisServer = ":6379"
			}
			c, err := redis.Dial("tcp", redisServer)
			if err != nil {
				log.Println("Error connecting to Redis: " + err.Error() + " - Falling back to local cache")
				return stupidCacheConn{
					AllowedKeys: []string{"hubwayData"},
				}, nil
			}
			password := os.Getenv("REDIS_PASSWORD")
			if len(password) > 0 {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
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
	if len(httpPort) == 0 {
		httpPort = "80"
	}
	if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
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
	stations, err := getHubwayData()
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
