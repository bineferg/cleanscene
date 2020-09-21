package airports

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/cleanscene.flights/lib/google"
)

type Airports interface {
	AirCodeByCity(string, string) (string, error)
	FindClosestAirport(string) (Edge, error)
}

func New(fname, edgeKey string, googleApi google.Places) (Airports, error) {
	var as service
	var airports = make(map[string]string)
	file, err := os.Open(fname)
	if err != nil {
		return as, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		arr := strings.Split(scanner.Text(), " - ")
		airports[arr[0]] = strings.Join(arr[1:], " ")
	}

	if err := scanner.Err(); err != nil {
		return as, err
	}
	as.cache = airports
	as.googleapi = googleApi
	as.edgeHost = "http://aviation-edge.com/v2/public/nearby?key="
	as.edgeKey = edgeKey
	return as, nil
}

type service struct {
	googleapi google.Places

	// Local datastore downloaded from https://datahub.io/core/airport-codes
	cache    map[string]string
	edgeHost string
	edgeKey  string
}

func isUrl(location string) bool { return strings.Contains(location, "http") }

func (as service) FindClosestAirport(location string) (Edge, error) {
	var (
		edges    []Edge
		lng, lat float64
		err      error
	)

	if isUrl(location) {
		coords := strings.Split(location, "?q=")
		lng, lat, err = as.googleapi.QueryCoordinates(coords[1], google.CoordQuery)

	} else {
		lng, lat, err = as.googleapi.QueryCoordinates(location, google.TextQuery)
	}
	if err != nil {
		return Edge{}, err
	}
	edges, err = as.nearestAirportByCoords(lng, lat)
	if err != nil {
		return Edge{}, err
	}
	for _, edge := range edges {
		if _, ok := as.cache[edge.Code]; ok {
			return edge, nil
		}
	}
	return Edge{}, errors.New("NoAirportFound")
}

func (as service) AirCodeByCity(city, country string) (string, error) {
	var aircode string
	for code, place := range as.cache {
		if strings.Contains(place, strings.ToUpper(city)) {
			aircode = code
		}
		if strings.Contains(place, strings.ToUpper(country)) {
			aircode = code

		}
		if strings.Contains(place, strings.ToUpper(city)) && strings.Contains(place, strings.ToUpper(country)) {
			return aircode, nil
		}
	}
	return aircode, nil

}

type Edge struct {
	Code     string `json:"codeIataAirport"`
	Country  string `json:"nameCountry"`
	CityCode string `json:"codeIataCity"`
}

type Edges []Edge

func (as service) nearestAirportByCoords(lng, lat float64) (Edges, error) {
	var eResp = Edges{}
	query := fmt.Sprintf("%s%s&lat=%f&lng=%f&distance=500", as.edgeHost, as.edgeKey, lat, lng)
	resp, err := http.Get(query)
	if err != nil {
		return eResp, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&eResp)
	if len(eResp) == 0 {
		return eResp, errors.New("No airports found!")
	}
	return eResp, nil
}
